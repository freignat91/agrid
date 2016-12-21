package agridapi

import (
	"fmt"
	"github.com/freignat91/agrid/server/gnode"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"sync"
	"time"
)

type gnodeClient struct {
	api      *AgridAPI
	id       string
	client   gnode.GNodeServiceClient
	nodeName string
	nodeHost string
	ctx      context.Context
	stream   gnode.GNodeService_GetClientStreamClient
	recvChan chan *gnode.AntMes
	lock     sync.RWMutex
	conn     *grpc.ClientConn
	nbNode   int
}

func (g *gnodeClient) init(api *AgridAPI) error {
	g.api = api
	g.ctx = context.Background()
	g.recvChan = make(chan *gnode.AntMes)
	if err := g.connectServer(); err != nil {
		return err
	}
	if err := g.startServerReader(); err != nil {
		return err
	}
	api.info("Client %s connected to node %s (%s)\n", g.id, g.nodeName, g.nodeHost)
	return nil
}

func (g *gnodeClient) connectServer() error {
	cn, err := grpc.Dial(g.api.serverAddress,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second*20))
	if err != nil {
		return err
	}
	g.conn = cn
	g.client = gnode.NewGNodeServiceClient(g.conn)
	ret, errp := g.client.Ping(g.ctx, &gnode.AntMes{})
	if errp != nil {
		return errp
	}
	g.nodeName = ret.Name
	g.nodeHost = ret.Host
	g.nbNode = int(ret.NbNode)
	return nil
}

func (g *gnodeClient) startServerReader() error {
	stream, err := g.client.GetClientStream(g.ctx)
	if err != nil {
		return err
	}
	g.stream = stream
	ack, err2 := g.stream.Recv()
	if err2 != nil {
		g.api.info("Client register EOF\n")
		close(g.recvChan)
		return fmt.Errorf("Client register error: %v\n", err2)
	}
	g.id = ack.FromClient
	g.api.info("Client register: %s\n", g.id)
	go func() {
		for {
			mes, err := g.stream.Recv()
			if err == io.EOF {
				g.api.debug("Server stream EOF\n")
				close(g.recvChan)
				return
			}
			if err != nil {
				g.api.error("Server stream error: %v\n", err)
				return
			}
			if mes.NoBlocking {
				select {
				case g.recvChan <- mes:
					//fmt.Printf("receive mes noBlocking: %v\n", mes)
				default:
					//fmt.Printf("receive mes noBlocking (wipeout): %v\n", mes)
				}
			} else {
				//fmt.Printf("receive mes Blocking: %v\n", mes)
				g.recvChan <- mes
			}
			g.api.debug("Receive answer: %v\n", mes)
		}
	}()
	return nil
}

func (g *gnodeClient) createMessage(target string, returnAnswer bool, functionName string, args ...string) *gnode.AntMes {
	return gnode.NewAntMes(target, returnAnswer, functionName, args...)
}

func (g *gnodeClient) createSendMessageNoAnswer(target string, functionName string, args ...string) error {
	mes := gnode.NewAntMes(target, false, functionName, args...)
	_, err := g.sendMessage(mes, true)
	return err
}

func (g *gnodeClient) createSendMessage(target string, waitForAnswer bool, functionName string, args ...string) (*gnode.AntMes, error) {
	mes := gnode.NewAntMes(target, true, functionName, args...)
	return g.sendMessage(mes, waitForAnswer)
}

func (g *gnodeClient) sendMessage(mes *gnode.AntMes, wait bool) (*gnode.AntMes, error) {
	g.lock.Lock()
	defer g.lock.Unlock()
	mes.FromClient = g.id
	//fmt.Printf("Order: %d size: %d\n", mes.Order, len(mes.Data))
	err := g.stream.Send(mes)
	if err != nil {
		return nil, err
	}
	//g.printf(Info, "Message sent: %v\n", mes)
	if wait {
		ret := <-g.recvChan
		return ret, nil
	}
	return nil, nil
}

func (g *gnodeClient) getNextAnswer(timeout int) (*gnode.AntMes, bool) {
	if timeout > 0 {
		timer := time.AfterFunc(time.Millisecond*time.Duration(timeout), func() {
			g.recvChan <- &gnode.AntMes{Function: "timeout"}
		})
		mes := <-g.recvChan
		timer.Stop()
		if mes == nil {
			return nil, false
		}
		if timeout > 0 && mes.Function == "timeout" {
			return nil, false
		}
		return mes, true
	}
	mes := <-g.recvChan
	return mes, true
}

func (g *gnodeClient) close() {
	if g.conn != nil {
		g.conn.Close()
	}
}
