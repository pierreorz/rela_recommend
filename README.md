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
go build
./rela_recommend --conf=conf.toml  # conf.toml需要从服务器上拷贝
```