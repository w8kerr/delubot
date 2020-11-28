/*
 * Models for Pixiv Fanbox objects
 */

package models

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

/*
 * Basic object models
 */

// FanboxUser A Fanbox user
type FanboxUser struct {
	UserID  string `json:"userId" bson:"userId"`
	Name    string `json:"name" bson:"name"`
	IconURL string `json:"iconUrl" bson:"iconUrl"`

	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"` // Not parsed from request, must be applied manually
}

// FanboxPost A Fanbox post
type FanboxPost struct {
	OID               bson.ObjectId          `json:"_id" bson:"_id,omitempty"`
	ID                string                 `json:"id" bson:"id"`
	Title             string                 `json:"title" bson:"title"`
	CoverImageURL     string                 `json:"coverImageUrl" bson:"coverImageUrl"`
	FeeRequired       int                    `json:"feeRequired" bson:"feeRequired"`
	PublishedDateTime time.Time              `json:"publishedDatetime" bson:"publishedDatetime"`
	UpdatedDateTime   time.Time              `json:"updatedDatetime" bson:"updatedDatetime"`
	Type              string                 `json:"type" bson:"type"`
	Body              map[string]interface{} `json:"body" bson:"body"`
	Tags              []string               `json:"tags" bson:"tags"`
	Excerpt           string                 `json:"excerpt" bson:"excerpt"`
	IsLiked           bool                   `json:"isLiked" bson:"isLiked"`
	LikeCount         int                    `json:"likeCount" bson:"likeCount"`
	CommentCount      int                    `json:"commentCount" bson:"commentCount"`
	RestrictedFor     int                    `json:"restrictedFor" bson:"restrictedFor"`
	User              FanboxUser             `json:"user" bson:"user"`
	CreatorID         string                 `json:"creatorId" bson:"creatorId"`
	HasAdultContent   bool                   `json:"hasAdultContent" bson:"hasAdultContent"`
	Status            string                 `json:"published" bson:"published"`
}

// FanboxComment A top level Fanbox comment
type FanboxComment struct {
	OID             bson.ObjectId   `json:"_id" bson:"_id,omitempty"`
	PostID          string          `json:"post_id" bson:"post_id"` // Not parsed from request, must be applied manually
	ID              string          `json:"id" bson:"id"`
	ParentCommentID string          `json:"parentCommentId" bson:"parentCommentId"`
	RootCommentID   string          `json:"rootCommentId" bson:"rootCommentId"`
	Body            string          `json:"body" bson:"body"`
	CreatedDateTime time.Time       `json:"createdDatetime" bson:"createdDatetime"`
	LikeCount       int             `json:"likeCount" bson:"likeCount"`
	IsLiked         bool            `json:"isLiked" bson:"isLiked"`
	IsOwn           bool            `json:"isOwn" bson:"isOwn"`
	User            FanboxUser      `json:"user" bson:"user"`
	Replies         []FanboxComment `json:"replies" bson:"replies"`
	WasReply        bool            `json:"was_reply" bson:"was_reply"` // Not parsed from request, used to flatten replies
	Deleted         bool            `json:"deleted" bson:"deleted"`     // Not parsed from request, must be applied manually
}

// SameAs Check if two comments are the same (by Post ID and Comment ID)
func (cmnt *FanboxComment) SameAs(comp FanboxComment) bool {
	return cmnt.PostID == comp.PostID && cmnt.ID == comp.ID
}

// MadeBefore Check if the receiver comment was made before the parameter comment
func (cmnt *FanboxComment) MadeBefore(comp FanboxComment) bool {
	return cmnt.CreatedDateTime.Before(comp.CreatedDateTime)
}

/*
 * HTTP request models
 */

// FanboxResponse Response carrying Fanbox objects
type FanboxResponse struct {
	Body struct {
		Items   []map[string]interface{} `json:"items"`
		NextURL string                   `json:"nextUrl"`
	} `json:"body"`
}
