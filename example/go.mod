module github.com/pubgo/protobuild/example

go 1.18

require (
	github.com/pubgo/protobuild v0.0.6
	github.com/pubgo/xerror v0.4.18
	google.golang.org/protobuf v1.28.0
	gorm.io/driver/sqlite v1.3.2
	gorm.io/gorm v1.23.5
)

replace github.com/pubgo/protobuild v0.0.6 => ../

require (
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.12 // indirect
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
)
