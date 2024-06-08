### 本地运行
```bash
git clone git@e.coding.net:rela/rela_recommend.git
cd rela_recommend
go run rela_recommend.go --conf=conf/conf.toml
```

### 发布
```bash
git clone git@coding.net:rela/rela_recommend.git
cd rela_recommend
make build
# 将algo_files目录copy到工作目录
./rela_recommend --conf=conf.toml  # conf.toml需要从服务器上拷贝
```

### 缓存
```
缓存使用二级缓存来节省内存空间，持久化在pika内，缓存在redis
缓存支持压缩格式
| key后缀 | 压缩解析方式 |
| --- | --- |
| .gz | gzip压缩 |
| .gzip | gzip压缩 |

```

### match 速配

| 时间 | 版本 | 负责人 | 算法 | 特征 | 目标 | 说明 |
| --- | --- | --- | --- | --- | --- | --- |
| 2018-11-02 | v1.0 | arvin | tree | 离线采集基本信息特征 | 是否喜欢 |  |

### live 直播

| 时间 | 版本 | 负责人 | 算法 | 特征 | 目标 | 说明 |
| --- | --- | --- | --- | --- | --- | --- |
| 2019-03-01 | v1.0 | arvin | gbdt+lr | 离线采集基本信息特征 | 观看大于3分钟 |  |
| 2019-03-15 | v1.1 | arvin | gbdt+lr | 线上采集基本信息特征 | 观看大于3分钟 |  |
| 2019-03-30 | v1.2 | arvin | gbdt+lr | 线上采集基本信息特征 + view5的embedding | 观看大于5分钟 |  |


### 
