#! /bin/bash
echo '[pipeline 循环执行 100 万次]'
go test -bench='(Concurrent_32|Serial_32|Complex_6|Connect_8x8)$' -benchtime=1000000x -benchmem -cpu=8