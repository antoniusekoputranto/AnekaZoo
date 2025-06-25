## **AnekaZoo Animal API Application Operations Guide**

This is a Go (Golang) REST API application for managing AnekaZoo animal data, fulfilling the task requirements.

### **Project Structure**

.  
├── go.mod          \# Go module definition and dependencies  
├── go.sum          \# Cryptographic checksums of dependencies  
├── main.go         \# Main API application logic  
└── README.md       \# This document

**Contents of go.mod:**

module AnekaZoo

go 1.24.4

require github.com/gorilla/mux v1.8.1

**Contents of go.sum:**

github.com/gorilla/mux v1.8.1 h1:TuBL49tXwgrFYWhqrNgrUNEY92u81SPhu7sTdzQEiWY=  
github.com/gorilla/mux v1.8.1/go.mod h1:AKf9I4AEqPTmMytcMc0KkNouC66V3BtZ4qD5fmWSiMQ=

### **Storage System**

For simplicity and in line with the flexibility mentioned in the task, this application uses **in-memory storage**. This means that all animal data will be lost every time the application is stopped and restarted.

### **How to Run the Application**

#### **Running the Application**

Ensure you have Go installed on your system (version 1.16 or newer recommended).

1. **Clone this repository** (if this is a Git repository).  
2. **Navigate to the project directory**:  
   cd \<project\_directory\_name\>

3. **Download necessary modules** (only gorilla/mux):  
   go mod tidy

4. **Run the application**:  
   go run main.go

   The application will start running on port 8000\. You will see the following output in the console:  
   Starting server at port 8000

### **API Addresses**

The application exposes API endpoints at http://localhost:8000 with a /v1 version prefix.

Here are the available endpoints:

* **GET /v1/animals**  
  * Retrieves a list of all existing animals.  
  * **Response:** 200 OK with an array of animal objects, or 404 Not Found if no animals are found.  
* **GET /v1/animals/{id}**  
  * Retrieves details of an animal by its ID.  
  * **Response:** 200 OK with the animal object, or 404 Not Found if the animal is not found.  
* **POST /v1/animals**  
  * Creates a new animal entry.  
  * **Example Payload (Request Body):::**  
    {  
      "id": 101,  
      "name": "panda",  
      "class": "mammal",  
      "legs": 4  
    }

  * **Response:** 201 Created with the created animal object on success.  
  * **Errors:** 400 Bad Request if the request body is invalid or ID is not provided. 409 Conflict if an animal with the same ID already exists.  
* **PUT /v1/animals/{id}**  
  * Updates an existing animal or creates a new animal if the ID does not exist (upsert operation).  
  * **Example Payload (Request Body):::**  
    {  
      "name": "grizzly bear",  
      "class": "mammal",  
      "legs": 4  
    }

    (Note that the id in the body is ignored; the id from the path parameter will be used.)  
  * **Response:** 200 OK with the updated animal object if successfully updated. 201 Created with the created animal object if the ID did not exist previously.  
  * **Errors:** 400 Bad Request if the ID in the path is invalid or the body request is invalid.  
* **DELETE /v1/animals/{id}**  
  * Deletes an animal by its ID.  
  * **Response:** 204 No Content on successful deletion.  
  * **Errors:** 404 Not Found if the animal is not found.