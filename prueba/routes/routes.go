package routes

import (
	"github.com/FelipeMarchantVargas/Prueba/controllers"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, uc *controllers.UserController) {

	// User routes
	//app.Get("/api/user/:id", uc.GetUser)
	app.Post("/api/login", uc.Login)
	app.Post("/api/register", uc.CreateUser)
	app.Delete("/api/user/:id", uc.DeleteUser)
	app.Get("/api/user", uc.User)
	app.Post("/api/logout", uc.Logout)

	// Themes and Comments routes
	app.Post("/api/CreateTheme", uc.CreateTheme)
	app.Get("/api/GetTheme", uc.GetThemes)
	app.Get("/api/GetTheme/:id", uc.GetTheme)
	app.Get("/api/GetThemesByUser", uc.GetThemesByUser)
	app.Delete("/api/DeleteTheme/:id", uc.DeleteTheme)
	app.Post("/api/CreateTComment/:id", uc.CreateComment)
	app.Get("/api/GetComments/:id", uc.GetComments)
	app.Get("/api/GetComment/:id", uc.GetComment)
	app.Delete("/api/DeleteComment/:id", uc.DeleteComment)

	// Archivos
	app.Post("/api/UploadFiles", uc.UploadFile)
	app.Get("/api/GetFiles", uc.GetUploadedFiles)
	app.Get("/api/Download/:id", uc.DownloadFile)
}
