#! /bin/bash
echo '[pipeline 并行执行 100 万次]'
go test -bench='Parallel$' -benchtime=1000000x -benchmem -cpu=1,4,8