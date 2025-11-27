package controller

import (
	"Tools3-Project/models"
	"context"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var eventCollection *mongo.Collection

func InitEventCollection(db *mongo.Database) {
	eventCollection = db.Collection("events")
}

// CREATE EVENT
func CreateEvent(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("user")
	if email == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	event.ID = primitive.NewObjectID()
	event.Organizer = email.(string)
	event.Attendees = []models.Attendee{
		{Email: email.(string), Status: "organizer"},
	}

	_, err := eventCollection.InsertOne(context.Background(), event)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(200, gin.H{"message": "Event created", "eventId": event.ID.Hex()})
}

// GET EVENTS ORGANIZED BY USER
func GetMyOrganizedEvents(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("user")

	if email == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	cursor, err := eventCollection.Find(context.Background(), bson.M{"organizer": email.(string)})
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}

	var events []models.Event
	if err := cursor.All(context.Background(), &events); err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse events"})
		return
	}

	c.JSON(200, events)
}

// GET EVENTS USER IS INVITED TO
func GetMyInvitedEvents(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("user")

	if email == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	cursor, err := eventCollection.Find(context.Background(), bson.M{"attendees.email": email.(string)})
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}

	var events []models.Event
	if err := cursor.All(context.Background(), &events); err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse events"})
		return
	}

	c.JSON(200, events)
}

// INVITE USER
func InviteUser(c *gin.Context) {
	var input models.InviteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	eventID, err := primitive.ObjectIDFromHex(input.EventID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid event ID"})
		return
	}

	_, err = eventCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": eventID},
		bson.M{"$addToSet": bson.M{"attendees": models.Attendee{Email: input.Email, Status: "invited"}}},
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to invite user"})
		return
	}

	c.JSON(200, gin.H{"message": "User invited"})
}

// RESPOND TO EVENT
func RespondToEvent(c *gin.Context) {
	var input models.ResponseInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	eventID, err := primitive.ObjectIDFromHex(input.EventID)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid event ID"})
		return
	}

	_, err = eventCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": eventID, "attendees.email": input.Email},
		bson.M{"$set": bson.M{"attendees.$.status": input.Status}},
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update response"})
		return
	}

	c.JSON(200, gin.H{"message": "Response recorded"})
}

// GET ATTENDEES
func GetEventAttendees(c *gin.Context) {
	eventID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	err = eventCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&event)
	if err != nil {
		c.JSON(404, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(200, event.Attendees)
}

//  EVENT 
func DeleteEvent(c *gin.Context) {
    session := sessions.Default(c)
    email := session.Get("user")

    id := c.Param("id")
    objID, _ := primitive.ObjectIDFromHex(id)

    res, err := eventCollection.DeleteOne(context.Background(),
        bson.M{"_id": objID, "organizer": email.(string)})

    if err != nil {
        c.JSON(500, gin.H{"error": "Database error"})
        return
    }

    if res.DeletedCount == 0 {
        c.JSON(403, gin.H{"error": "Not allowed"})
        return
    }

    c.JSON(200, gin.H{"message": "Event deleted"})
}

// SEARCH EVENTS 
func SearchEvents(c *gin.Context) {
    q := c.Query("q")

    filter := bson.M{
        "$or": []bson.M{
            {"title": bson.M{"$regex": q, "$options": "i"}},
            {"description": bson.M{"$regex": q, "$options": "i"}},
        },
    }

    cursor, _ := eventCollection.Find(context.Background(), filter)

    var events []models.Event
    cursor.All(context.Background(), &events)

    c.JSON(200, events)
}
