// API Documentation: https://documenter.getpostman.com/view/16354665/UV5RkKeFhttps://documenter.getpostman.com/view/16354665/UV5RkKeF
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"strconv"
    "golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	
)
func main() {
	connect()
	handleRequest()
}

var client *mongo.Client
func connect() {

	clientOptions := options.Client().ApplyURI("mongodb+srv://Promit_Revar:CaptainZaltan@cluster0.oxi5g.mongodb.net/Go_task?retryWrites=true&w=majority")
	client, _ = mongo.NewClient(clientOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.Background(), readpref.Primary())

	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	} else {
		log.Println("Connected to MondoDB Server")
	}

}

// Structures for User and posts 
type User struct{
	ID        string  `json:"id" bson:"id"`                              // User's Id 
	Name      string  `json:"name" bson:"name"`                         // Name of the user
	Email     string  `json:"email" bson:"email"`                      // email id of the user
	Password  string  `json:"pass" bson:"password"`                   // password filled that has to be hashed and salted before storing
}
type Post struct{
	UserID       string  `json:"uid" bson:"uid"`           
	ID           string  `json:"id" bson:"id"`
	Caption      string  `json:"caption" bson:"caption"`
	ImgUrl       string  `json:"url" bson:"url"`
	PostTime     time.Time `json:"posttime" bson:"posttime"`
}



// Routes Defined...
func handleRequest() {

	http.HandleFunc("/", homePage)                   // For home page...
	http.HandleFunc("/users", CreateUser)            // Creating user...
	http.HandleFunc("/users/", returnSingleUser)     // Searching user by providing id as parameter in request url "/users/<id>"
    http.HandleFunc("/posts", CreatePost)            // Route for creating a post...
	http.HandleFunc("/posts/", SearchPost)           // Route for Searching a post based on post id
	http.HandleFunc("/posts/users/", ReturnALLPosts) // Route for providing all the posts for a given user by taking UserId as 
	                                                 // request parameter "/posts/users/<UserId>"
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe", err)
	}
}
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello Promit!")
	
}


func CreatePost(response http.ResponseWriter, request *http.Request){
	request.ParseForm()
	decoder := json.NewDecoder(request.Body)
	var newPost Post
	newPost.PostTime = time.Now()
	err := decoder.Decode(&newPost)
	if err != nil {
		panic(err)
	}
	//log.Println(newPost.ID)
	
	insertPost(newPost)
}
// Create User function for creating new user..
func CreateUser(response http.ResponseWriter, request *http.Request) {

	
		request.ParseForm()
		decoder := json.NewDecoder(request.Body)
		var newUser User
		
		err := decoder.Decode(&newUser)
		if err != nil {
			panic(err)
		}
		
        newUser.Password=hashAndSalt([]byte(newUser.Password))
		
		insertUser(newUser)
	
}

// For querying Users on id
func returnSingleUser(response http.ResponseWriter, request *http.Request) {

	request.ParseForm()
	var id string = request.URL.Path
	
	id = id[7:]
	fmt.Println(id)
	var user User
	collection := client.Database("Go_task").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"id": id}).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	
	json.NewEncoder(response).Encode(user)
}

// Search post based on id
func SearchPost(response http.ResponseWriter, request *http.Request) {

	request.ParseForm()
	var id string = request.URL.Path
	
	id = id[7:]
	
	var post Post
	collection := client.Database("Go_task").Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"id": id}).Decode(&post)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	
	json.NewEncoder(response).Encode(post)
}
// Function that Returns all posts of a user
func ReturnALLPosts(response http.ResponseWriter, request *http.Request){
	var posts []Post
	    
	request.ParseForm()
	var u string = request.URL.Path
	query := request.URL.Query()
	index,_ := strconv.Atoi(query.Get("index"))    // Getting Cursor value from user to implement cursor paggination
	uid := u[13:]
	
 

		collection := client.Database("Go_task").Collection("posts")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cursor, err := collection.Find(ctx, bson.M{"uid":uid})
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var post Post
			cursor.Decode(&post)
		
			posts = append(posts, post)

		}
		
		if err = cursor.Err(); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		
		json.NewEncoder(response).Encode(posts[index:])
}
// function for inserting the User in the database
func insertUser(user User) {
	collection := client.Database("Go_task").Collection("users")
	insertResult, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted user with ID:", insertResult.InsertedID)
}
// Function to insert Post in database...
func insertPost(post Post) {
	collection := client.Database("Go_task").Collection("posts")
	insertResult, err := collection.InsertOne(context.TODO(), post)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted post with ID:", insertResult.InsertedID)
}
// Hash and Salt password before saving...
func hashAndSalt(pwd []byte) string {
    
   
    hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
    if err != nil {
        log.Println(err)
    }
    
    return string(hash)
}