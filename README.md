# project-bee

```shell
make run
make test
```
# ep5 区块头    

# ep6 验证区块
- 生成随机区块
- header 的 hash 方法
- 验证区块头 hash

# ep7 内存池
- blockchain.go 
    -- GetHeader() 增加 	bc.lock.Lock()
    -- Height() 增加 	bc.lock.RLock()
    -- addBlockWithoutValidation() 增加 bc.lock.Lock()

- 增加 txPool.go
    -- Add()
    -- Has()
    -- Len()
    -- Flush()
- 增加 txPool_test.go

- server.go 
  -- handleTransaction()

# ep8 内存池

- txPool.go
  -- NewTxMapSorter()  FIFO的tx排序方法
   
- txPool_test.go
   -- TestSortTransactions()

- encoding.go 注册elliptic.P256()椭圆曲线的签名算法
  -- Encode(tx *Transaction) 
  -- Decode(tx *Transaction)

# ep9 rpc

- rpc.go