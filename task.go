package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// Define a struct to represent a task
type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	DueDate     time.Time          `bson:"due_date" json:"due_date"`
	Status      string             `bson:"status" json:"status"`
}

// Create a new task
func createTask(c *gin.Context) {
	// Parse JSON payload into a Task struct
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		// Return a Bad Request response if JSON parsing fails
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a unique ID for the task
	task.ID = primitive.NewObjectID()

	// Store the task in the database
	collection := client.Database("taskdb").Collection("tasks")
	_, err := collection.InsertOne(context.Background(), task)
	if err != nil {
		// Return an Internal Server Error response if database operation fails
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Return the created task with a status code of 201 (Created)
	c.JSON(http.StatusCreated, task)
}

// Retrieve a task
func getTask(c *gin.Context) {
	// Retrieve the task ID from the URL parameter
	id := c.Param("id")

	// Convert the task ID string to a BSON ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// Return a Bad Request response if the ID is invalid
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Fetch the task from the database based on the ID
	collection := client.Database("taskdb").Collection("tasks")
	var task Task
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&task)
	if err != nil {
		// Return a Not Found response if task is not found
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Return the task details with a status code of 200 (OK)
	c.JSON(http.StatusOK, task)
}

// Update a task
func updateTask(c *gin.Context) {
	// Retrieve the task ID from the URL parameter
	id := c.Param("id")

	// Parse JSON payload into a Task struct
	var updatedTask Task
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		// Return a Bad Request response if JSON parsing fails
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the task in the database based on the ID
	collection := client.Database("taskdb").Collection("tasks")
	_, err := collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": updatedTask})
	if err != nil {
		// Return an Internal Server Error response if database operation fails
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Return the updated task details with a status code of 200 (OK)
	c.JSON(http.StatusOK, updatedTask)
}

// Delete a task
func deleteTask(c *gin.Context) {
	// Retrieve the task ID from the URL parameter
	id := c.Param("id")

	// Delete the task from the database based on the ID
	collection := client.Database("taskdb").Collection("tasks")
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		// Return an Internal Server Error response if database operation fails
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	// Return a success message with a status code of 200 (OK)
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// List all tasks
func listTasks(c *gin.Context) {
	// Fetch all tasks from the database
	collection := client.Database("taskdb").Collection("tasks")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		// Return an Internal Server Error response if database operation fails
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks"})
		return
	}
	defer cursor.Close(context.Background())

	var tasks []Task
	if err := cursor.All(context.Background(), &tasks); err != nil {
		// Return an Internal Server Error response if database operation fails
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks"})
		return
	}

	// Return the list of tasks with a status code of 200 (OK)
	c.JSON(http.StatusOK, tasks)
}

func main() {
	// Set up MongoDB client
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Gin router
	router := gin.Default()

	// Define API endpoints
	router.POST("/tasks", createTask)
	router.GET("/tasks/:id", getTask)
	router.PUT("/tasks/:id", updateTask)
	router.DELETE("/tasks/:id", deleteTask)
	router.GET("/tasks", listTasks)

	// Start the HTTP server on port 8080
	router.Run(":8080")
}
