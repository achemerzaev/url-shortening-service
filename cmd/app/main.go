// @title URL shortener API
// @version 1.0
// @description URL shortener service

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import "github.com/achemerzaev/url-shortening-service/internal/app"

// проверка билда с секретами и внутренней сетю, разделение конфигов и докер компоуза на девелопмент и релиз
// сделат мейн и хендлеры тонкими
// деплой, сертификат
// ci/cd

func main() {
	app.Run()
}
