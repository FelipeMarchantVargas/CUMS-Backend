package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/FelipeMarchantVargas/Prueba/models"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const SecretKey = "secret"

type UserController struct {
	client *mongo.Client
}

func NewUserController(client *mongo.Client) *UserController {
	return &UserController{client}
}

func (uc UserController) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	user := models.User{}

	collection := uc.client.Database("gomongodb").Collection("users")

	err1 := collection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&user)

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	uj, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
		return err
	}

	c.Set(fiber.HeaderContentType, "application/json")
	c.Status(http.StatusOK)
	c.Send(uj)

	return nil
}

func (uc UserController) CreateUser(c *fiber.Ctx) error {
	u := models.User{}

	body := c.Body()

	err := json.Unmarshal(body, &u)

	if err != nil {
		c.Status(http.StatusBadRequest)
		return err
	}

	u.Id = primitive.NewObjectID()

	u.CreationTime = time.Now()

	u.Student = true

	u.Ayudante = false

	password, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 14)

	u.Password = string(password)

	collection := uc.client.Database("gomongodb").Collection("users")

	_, err = collection.InsertOne(context.TODO(), u)

	if err != nil {
		c.Status(http.StatusInternalServerError)
		return err
	}

	uj, err := json.Marshal(u)

	if err != nil {
		c.Status(http.StatusInternalServerError)
		return err
	}

	c.Set(fiber.HeaderContentType, "application/json")
	c.Status(http.StatusCreated)
	c.Send(uj)

	return nil
}

func (uc UserController) DeleteUser(c *fiber.Ctx) error {

	id := c.Params("id")

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	collection := uc.client.Database("gomongodb").Collection("users")

	_, err1 := collection.DeleteOne(context.TODO(), bson.M{"_id": oid})

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	c.Status(http.StatusOK)
	fmt.Printf("Deleted user %s\n", oid)

	return nil
}

func (uc UserController) Login(c *fiber.Ctx) error {
	u := models.User{}
	user := models.User{}

	body := c.Body()
	err := json.Unmarshal(body, &u)

	if err != nil {
		c.Status(http.StatusBadRequest)
		return err
	}

	collection := uc.client.Database("gomongodb").Collection("users")

	err1 := collection.FindOne(context.TODO(), bson.M{"email": u.Email}).Decode(&user)

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return c.JSON(fiber.Map{
			"message": "Correo electrónico no encontrado",
		})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Contraseña incorrecta",
		})
	}

	expirationTime := time.Now().Add(time.Hour * 24).Unix()
	jwtTime := jwt.NewTime(float64(expirationTime))

	id := user.Id.Hex()

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    string(id),
		ExpiresAt: jwtTime,
	})

	token, err := claims.SignedString([]byte(SecretKey))

	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"message": "No se pudo iniciar sesión",
		})
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "Éxito",
	})
}

func (uc UserController) User(c *fiber.Ctx) error {

	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Not authenticated",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	user := models.User{}

	collection := uc.client.Database("gomongodb").Collection("users")

	oid, _ := primitive.ObjectIDFromHex(claims.Issuer)

	err1 := collection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&user)

	if err1 != nil {
		return err1
	}

	return c.JSON(user)
}

func (uc UserController) Logout(c *fiber.Ctx) error {

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "Exito",
	})
}

//Functions CRUD blog

func (uc UserController) CreateTheme(c *fiber.Ctx) error {

	//GET user session
	cookie := c.Cookies("jwt")

	token, err0 := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err0 != nil {
		fmt.Printf("Error: %s\n", err0)
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Not authenticated",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	user := models.User{}

	collection := uc.client.Database("gomongodb").Collection("users")

	oid, _ := primitive.ObjectIDFromHex(claims.Issuer)

	err1 := collection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&user)

	if err1 != nil {
		return err1
	}
	//Insert blog in the DB.

	blog := models.BlogTheme{}

	body := c.Body()

	err2 := json.Unmarshal(body, &blog)

	if err2 != nil {
		c.Status(http.StatusBadRequest)
		return err2
	}

	blog.IdTheme = primitive.NewObjectID()

	blog.IdCreator = user.Id

	blog.NameCreator = user.Name

	blog.CreationTime = time.Now()

	collection2 := uc.client.Database("gomongodb").Collection("blog")

	_, err2 = collection2.InsertOne(context.TODO(), blog)

	if err2 != nil {
		c.Status(http.StatusInternalServerError)
		fmt.Print("error 2\n")
		return err2
	}

	uj, err2 := json.Marshal(blog)

	if err2 != nil {
		c.Status(http.StatusInternalServerError)
		fmt.Print("error 3\n")
		return err2
	}

	c.Set(fiber.HeaderContentType, "application/json")
	c.Status(http.StatusCreated)
	c.Send(uj)

	return c.JSON(fiber.Map{
		"message": "Exito",
	})
}

func (uc UserController) DeleteTheme(c *fiber.Ctx) error {
	id := c.Params("id")

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	collection := uc.client.Database("gomongodb").Collection("blog")

	_, err1 := collection.DeleteOne(context.TODO(), bson.M{"idBlog": oid})

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	c.Status(http.StatusOK)
	fmt.Printf("Deleted blog %s\n", oid)

	return nil
}

func (uc UserController) GetThemes(c *fiber.Ctx) error {

	var blogs []models.BlogTheme

	collection := uc.client.Database("gomongodb").Collection("blog")

	opts := options.Find().SetSort(bson.D{{Key: "creationDate", Value: -1}})

	results, err1 := collection.Find(context.TODO(), bson.M{}, opts)

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return nil
	}
	for results.Next(context.TODO()) {
		var blog models.BlogTheme
		results.Decode(&blog)
		blogs = append(blogs, blog)
	}

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	uj, err := json.Marshal(blogs)
	if err != nil {
		fmt.Println(err)
		return err
	}

	c.Set(fiber.HeaderContentType, "application/json")
	c.Status(http.StatusOK)
	c.Send(uj)

	return nil

}

func (uc UserController) GetTheme(c *fiber.Ctx) error {
	id := c.Params("id")

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	blog := models.BlogTheme{}

	collection := uc.client.Database("gomongodb").Collection("blog")

	err1 := collection.FindOne(context.TODO(), bson.M{"idBlog": oid}).Decode(&blog)

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	uj, err := json.Marshal(blog)
	if err != nil {
		fmt.Println(err)
		return err
	}

	c.Set(fiber.HeaderContentType, "application/json")
	c.Status(http.StatusOK)
	c.Send(uj)

	return nil

}

func (uc UserController) GetThemesByUser(c *fiber.Ctx) error {
	var blogs []models.BlogTheme

	//GET user session
	cookie := c.Cookies("jwt")

	token, err0 := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err0 != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Not authenticated",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	user := models.User{}

	collection := uc.client.Database("gomongodb").Collection("users")

	oid, _ := primitive.ObjectIDFromHex(claims.Issuer)

	err1 := collection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&user)

	if err1 != nil {
		return err1
	}

	collection2 := uc.client.Database("gomongodb").Collection("blog")

	opts := options.Find().SetSort(bson.D{{Key: "creationDate", Value: -1}})

	results, err1 := collection2.Find(context.TODO(), bson.M{"idCreator": oid}, opts)

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	for results.Next(context.TODO()) {
		var blog models.BlogTheme
		results.Decode(&blog)
		blogs = append(blogs, blog)
	}

	uj, err := json.Marshal(blogs)
	if err != nil {
		fmt.Println(err)
		return err
	}

	c.Set(fiber.HeaderContentType, "application/json")
	c.Status(http.StatusOK)
	c.Send(uj)

	return nil
}

func (uc UserController) GetComments(c *fiber.Ctx) error {

	id := c.Params("id")

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	var comments []models.BlogComment

	collection := uc.client.Database("gomongodb").Collection("comment")

	results, err1 := collection.Find(context.TODO(), bson.M{"idTheme": oid})

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return nil
	}
	for results.Next(context.TODO()) {
		var comment models.BlogComment
		results.Decode(&comment)
		comments = append(comments, comment)
	}

	uj, err := json.Marshal(comments)
	if err != nil {
		fmt.Println(err)
		return err
	}

	c.Set(fiber.HeaderContentType, "application/json")
	c.Status(http.StatusOK)
	c.Send(uj)

	return nil
}

func (uc UserController) CreateComment(c *fiber.Ctx) error {
	id := c.Params("id")

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	//GET user session
	cookie := c.Cookies("jwt")

	token, err0 := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err0 != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Not authenticated",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	user := models.User{}

	collection := uc.client.Database("gomongodb").Collection("users")

	oid2, _ := primitive.ObjectIDFromHex(claims.Issuer)

	err1 := collection.FindOne(context.TODO(), bson.M{"_id": oid2}).Decode(&user)

	if err1 != nil {
		return err1
	}
	//Insert comment in the DB.

	comment := models.BlogComment{}

	body := c.Body()

	err2 := json.Unmarshal(body, &comment)

	if err2 != nil {
		c.Status(http.StatusBadRequest)
		return err2
	}

	comment.IdComment = primitive.NewObjectID()

	comment.IdCreator = user.Id

	comment.NameCreator = user.Name

	comment.IdTheme = oid

	comment.CreationTime = time.Now()

	collection2 := uc.client.Database("gomongodb").Collection("comment")

	_, err2 = collection2.InsertOne(context.TODO(), comment)

	if err2 != nil {
		c.Status(http.StatusInternalServerError)
		fmt.Print("error 2\n")
		return err2
	}

	uj, err2 := json.Marshal(comment)

	if err2 != nil {
		c.Status(http.StatusInternalServerError)
		fmt.Print("error 3\n")
		return err2
	}

	c.Set(fiber.HeaderContentType, "application/json")
	c.Status(http.StatusCreated)
	c.Send(uj)

	return c.JSON(fiber.Map{
		"message": "Exito",
	})
}

func (uc UserController) DeleteComment(c *fiber.Ctx) error {
	id := c.Params("id")

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	collection := uc.client.Database("gomongodb").Collection("comment")

	_, err1 := collection.DeleteOne(context.TODO(), bson.M{"idComment": oid})

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	c.Status(http.StatusOK)
	fmt.Printf("Deleted comment %s\n", oid)

	return nil
}

func (uc UserController) GetComment(c *fiber.Ctx) error {
	id := c.Params("id")

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	blog := models.BlogComment{}

	collection := uc.client.Database("gomongodb").Collection("comment")

	err1 := collection.FindOne(context.TODO(), bson.M{"idComment": oid}).Decode(&blog)

	if err1 != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	uj, err := json.Marshal(blog)
	if err != nil {
		fmt.Println(err)
		return err
	}

	c.Set(fiber.HeaderContentType, "application/json")
	c.Status(http.StatusOK)
	c.Send(uj)

	return nil

}

// Archivos
func (uc UserController) UploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("upload")

	if err != nil {
		return err
	}

	// Abrir el archivo subido
	uploadedFile, err := file.Open()
	if err != nil {
		return err
	}
	defer uploadedFile.Close()

	// Leer el contenido del archivo en memoria
	fileData, err := ioutil.ReadAll(uploadedFile)
	if err != nil {
		return err
	}

	// Guardar el archivo en MongoDB
	collection := uc.client.Database("gomongodb").Collection("uploadfiles")
	_, err = collection.InsertOne(context.Background(), bson.M{
		"filename": file.Filename,
		"data":     fileData,
	})
	if err != nil {
		return err
	}

	return c.SendString("Archivo subido y guardado en MongoDB correctamente.")
}

func (uc UserController) GetUploadedFiles(c *fiber.Ctx) error {
	// Conecta con la colección de MongoDB
	collection := uc.client.Database("gomongodb").Collection("uploadfiles")

	// Realiza una consulta a la colección
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	var files []map[string]interface{}

	// Itera a través de los resultados de la consulta
	for cursor.Next(context.Background()) {
		var fileData bson.M
		if err := cursor.Decode(&fileData); err != nil {
			return err
		}

		// Agrega los datos necesarios para mostrar el archivo (por ejemplo, el nombre y el ID)
		file := map[string]interface{}{
			"filename": fileData["filename"],
			"id":       fileData["_id"],
		}

		files = append(files, file)
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return c.JSON(files)
}

func (uc UserController) DownloadFile(c *fiber.Ctx) error {
	// Obtén el ID del archivo desde los parámetros de la URL
	fileID := c.Params("id")

	// Conecta con la colección de MongoDB
	collection := uc.client.Database("gomongodb").Collection("uploadfiles")

	// Busca el archivo en la colección por su ID
	var fileData bson.M
	if err := collection.FindOne(context.Background(), bson.M{"_id": fileID}).Decode(&fileData); err != nil {
		return err
	}

	// Extrae los datos del archivo
	filename := fileData["filename"].(string)
	data := fileData["data"].(string) // Asumiendo que los datos se almacenan como una cadena base64

	// Convierte los datos de base64 nuevamente en un archivo binario
	fileBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}

	// Configura la respuesta para la descarga
	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "application/octet-stream")

	// Escribe los datos del archivo en la respuesta
	c.Write(fileBytes)

	return nil
}
