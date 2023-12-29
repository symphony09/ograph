#! /bin/bash
echo '[pipeline 循环执行 100 万次]'
go test -bench='(Concurrent_32|Serial_32|Complex_6)$' -benchtime=1000000x -benchmem -cpu=1,4,8