# Rate Limiter

## description
---

實作一個middleware
* 限制每小時來自同一個 IP 的請求數量不得超過 1000
* 在 response headers 中加入剩餘的請求數量 (X-RateLimit-Remaining) 以及 rate limit 歸零的時間 (X-RateLimit-Reset)
* 如果超過限制的話就回傳 429 (Too Many Requests)
---
## TL;DR

* uses **gin** middleware and **redis** to record the remain requests and reset time.
---
## 使用的技術

* language：Go
* DB：redis
---
## 實作的過程

1. 怎麼模擬不同用戶的request，先搞定一個ip就好，略
2. 先弄成handler的形式，再弄成middleware
3. 設置redis
    * 安裝
        * `brew install redis`
        * `redis-server`
        * `redis-cli`
        * `go get -u github.com/go-redis/redis`
    * client的ip address => use c.ClientIP()
    * 模擬不同ip address的request => `curl --header "X-Forwarded-For: 1.2.3.4" http://localhost:8080/ping` 看起來不可行
    * 紀錄{ip address, remainRequest} pair => 
        * ERR invalid expire time in set => time.Duration(requestResetTime)*time.Second
4. code
    * 如果該ip是第一次訪問 => set redis {ip, requestLimit-1, requestResetTime} 
    * 否則 => decr value and check remainRequestInt < 0 ?
    * requestLimit減一後要重新取得值 => remainRequest, err = redisClient.Get(c.ClientIP()).Result()
    * 時間轉換：有點煩，以int64的型態把現在時間加上剩餘reset時間，再用time.Unix轉為time.Time型態即可。
5. middleware
    * middleware代表的涵意跟原本直接寫handler不太相同啊...
    * https://ithelp.ithome.com.tw/articles/10243831
    * 看完上面這篇後大概知道middleware在幹嘛了，
        * c.Next()會跳回原本的function，等原本的function都執行完才會回來(如果原本的c.Next()後面還有程式的話)
        * c.Abort()則不會，通常就在後面加上c.JSON()了
    * 是在response headers加入資訊...

