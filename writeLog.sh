#!/bin/bash

IFS='
'
for ll in `docker ps | grep agrid`;
do
        ee=${ll/%\ */}
        docker logs $ee >& ./logs/$ee.log
done 