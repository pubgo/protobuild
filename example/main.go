package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"
	
	example "github.com/pubgo/protobuild/example/proto"
	"github.com/pubgo/protobuild/ormpb"
	"github.com/pubgo/xerror"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gl "gorm.io/gorm/logger"
)

func main() {
	db := xerror.PanicErr(gorm.Open(sqlite.Open("./sqlite.db"), &gorm.Config{})).(*gorm.DB)
	db.Logger = gl.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gl.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gl.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})
	xerror.Panic(db.Callback().Create().Before("gorm:create").Register("pb:create", func(db *gorm.DB) {
		if db.Error != nil || db.Statement.Schema == nil || db.Statement.SkipHooks || db.Statement.Schema.BeforeSave || db.Statement.Schema.BeforeCreate {
			return
		}

		for _, field := range db.Statement.Schema.Fields {
			if field.FieldType != ormpb.Type {
				continue
			}

			if field.Name == "CreatedAt" || field.Name == "UpdatedAt" {
				field.ReflectValueOf(db.Statement.Context, db.Statement.ReflectValue).Set(reflect.ValueOf(ormpb.Now()))
			}
		}
	}))

	xerror.Panic(db.Callback().Update().Before("gorm:update").Register("pb:update", func(db *gorm.DB) {
		if db.Error != nil || db.Statement.Schema == nil || db.Statement.SkipHooks || db.Statement.Schema.BeforeSave || db.Statement.Schema.BeforeCreate {
			return
		}

		for _, field := range db.Statement.Schema.Fields {
			if field.FieldType != ormpb.Type {
				continue
			}

			if field.Name == "UpdatedAt" {
				field.ReflectValueOf(db.Statement.Context, db.Statement.ReflectValue).Set(reflect.ValueOf(ormpb.Now()))
			}
		}
	}))

	xerror.Panic(db.Callback().Delete().Before("gorm:delete").Register("pb:delete", func(db *gorm.DB) {
		if db.Error != nil || db.Statement.Schema == nil || db.Statement.SkipHooks || db.Statement.Schema.BeforeSave || db.Statement.Schema.BeforeCreate {
			return
		}

		for _, field := range db.Statement.Schema.Fields {
			if field.FieldType != ormpb.Type {
				continue
			}

			if field.Name == "DeletedAt" {
				field.ReflectValueOf(db.Statement.Context, db.Statement.ReflectValue).Set(reflect.ValueOf(ormpb.Now()))
			}
		}
	}))

	xerror.Panic(db.AutoMigrate(&example.Role{}))
	xerror.Panic(db.Save(&example.Role{}).Error)
	xerror.Panic(db.Save(&example.Role{}).Error)

	var m = &example.Role{}
	xerror.Panic(db.Save(m).Error)
	time.Sleep(time.Second)
	xerror.Panic(db.Updates(m).Error)
	fmt.Println(m.Id)
	xerror.Panic(db.First(m).Error)
	fmt.Println(m.Id)
	m.Id -= 1
	xerror.Panic(db.Delete(m).Error)
}
