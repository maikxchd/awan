package main

import (
	"os"

	"github.com/awanio/awan/internal/runtime"
	"github.com/awanio/awan/internal/user"
	"github.com/awanio/awan/pkg/helper"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

func newApp() *iris.Application {

	app := iris.Default()
	app.Logger().SetLevel("debug")

	runtime.Setup()

	if runtime.DBerror != nil {
		app.Logger().Fatalf("connect to sqlite3 failed")
		return nil
	}

	iris.RegisterOnInterrupt(func() {
		defer runtime.SQLDB.Close()
	})

	err := runtime.DB.AutoMigrate(&user.Users{}, &user.Credentials{})

	if err != nil {
		app.Logger().Fatalf(err.Error())
		return nil
	}

	app.OnErrorCode(iris.StatusNotFound, func(ctx iris.Context) {
		ctx.HTML("<b>Resource Not found</b>")
		// ctx.ServeFile("../../web/public/index.html")
	})

	api := app.Party("/api")
	{
		mvc.New(api.Party("/signup")).Register(runtime.DB).Handle(new(user.Signup))
		mvc.New(api.Party("/signin")).Register(runtime.DB).Handle(new(user.Signin))
		mvc.New(api.Party("/users")).Register(runtime.DB).Handle(new(user.Controller))
		mvc.New(api.Party("/apps")).Register(runtime.DB).Handle(new(user.Controller))
		mvc.New(api.Party("/resources")).Register(runtime.DB).Handle(new(user.Controller))
		mvc.New(api.Party("/teams")).Register(runtime.DB).Handle(new(user.Controller))
	}

	// app.Get("/{p:path}", func(ctx iris.Context) {
	// 	ctx.ServeFile("../../web/public/index.html")
	// })

	app.HandleDir("/", iris.Dir(helper.FromBasepath("web/public")), iris.DirOptions{IndexName: "index.html"})

	return app

}

func main() {
	app := newApp()
	port := ":8081"
	if s := os.Getenv("PORT"); s != "" {
		port = ":" + s
	}

	app.Listen(port)
}
