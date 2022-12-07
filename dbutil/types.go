package dbutil

import (
	"github.com/eth-library/dataset-dj/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Source contains information about the origin of the data contained within a MetaArchive
type Source struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Organisation string `json:"organisation"`
	Type         string `json:"type"`
}

type Order struct {
	OrderID   string   `json:"orderID" bson:"_id,omitempty"`
	ArchiveID string   `json:"archiveID" bson:"archiveID"`
	Email     string   `json:"email" bson:"email"`
	Date      string   `json:"date" bson:"date"`
	Status    string   `json:"status" bson:"status"`
	Sources   []Source `json:"sources" bson:"sources"`
}

type FileGroup struct {
	Source Source   `json:"source"`
	Files  util.Set `json:"files"`
}

func FileGroupToDB(fg FileGroup) FileGroupDB {
	return FileGroupDB{
		Source: fg.Source,
		Files:  fg.Files.ToSlice(),
	}
}

type FileGroupDB struct {
	Source Source   `json:"source"`
	Files  []string `json:"files"`
}

func DBToFileGroup(db FileGroupDB) FileGroup {
	return FileGroup{
		Source: db.Source,
		Files:  util.SetFromSlice(db.Files),
	}
}

func Union(fgs1 []FileGroup, fgs2 []FileGroup) ([]FileGroup, []Source) {
	var contentMap map[string]FileGroup
	for _, fgs := range [][]FileGroup{fgs1, fgs2} {
		contentMap = fillContentMap(fgs, contentMap)
	}
	var res []FileGroup
	var sources []Source
	for _, value := range contentMap {
		res = append(res, value)
		sources = append(sources, value.Source)
	}
	return res, sources
}

func Unify(fgs []FileGroup) ([]FileGroup, []Source) {
	var contentMap map[string]FileGroup
	contentMap = fillContentMap(fgs, contentMap)
	var res []FileGroup
	var sources []Source
	for _, value := range contentMap {
		res = append(res, value)
		sources = append(sources, value.Source)
	}
	return res, sources
}

func fillContentMap(fgs []FileGroup, contentMap map[string]FileGroup) map[string]FileGroup {
	for _, fg := range fgs {
		i, ok := contentMap[fg.Source.Name]
		if ok {
			contentMap[fg.Source.Name] = FileGroup{
				Source: i.Source,
				Files:  util.SetUnion(fg.Files, i.Files),
			}
		} else {
			contentMap[fg.Source.Name] = fg
		}
	}
	return contentMap
}

// MetaArchive is the blueprint for the zip archives that will be created once the user initiates
// the download process. Files is implemented as a set in order to avoid duplicate files within a
// metaArchive
type MetaArchive struct {
	ID          string      `json:"id"`
	Content     []FileGroup `json:"content"`
	Meta        string      `json:"meta"`
	TimeCreated string      `json:"timeCreated"`
	TimeUpdated string      `json:"timeUpdated"`
	Status      string      `json:"status"`
	Sources     []Source    `json:"sources"`
}

// Convert a MetaArchive to a MetaArchiveDB
func (arch MetaArchive) Convert() MetaArchiveDB {
	return MetaArchiveDB{
		ID:          arch.ID,
		Content:     util.Mapping(arch.Content, FileGroupToDB),
		Meta:        arch.Meta,
		TimeCreated: arch.TimeCreated,
		TimeUpdated: arch.TimeUpdated,
		Status:      arch.Status,
		Sources:     arch.Sources}
}

// ToBSON converts MetaArchive to binary JSON format
func (arch MetaArchive) ToBSON() bson.D {
	return bson.D{
		{
			"_id",
			arch.ID},
		{
			"content",
			util.Mapping(arch.Content, FileGroupToDB)},
		{
			"meta",
			arch.Meta},
		{
			"timeCreated",
			arch.TimeCreated},
		{
			"timeUpdated",
			arch.TimeUpdated},
		{
			"status",
			arch.Status},
		{
			"sources",
			arch.Sources}}
}

// MetaArchiveDB is a Wrapper type for MetaArchive as the custom type util.Set cannot be saved to the
// database
type MetaArchiveDB struct {
	ID          string        `json:"id"`
	Content     []FileGroupDB `json:"content"`
	Meta        string        `json:"meta"`
	TimeCreated string        `json:"timeCreated"`
	TimeUpdated string        `json:"timeUpdated"`
	Status      string        `json:"status"`
	Sources     []Source      `json:"sources"`
}

// Convert MetaArchiveDB to MetaArchive
func (arch MetaArchiveDB) Convert() MetaArchive {
	return MetaArchive{
		ID:          arch.ID,
		Content:     util.Mapping(arch.Content, DBToFileGroup),
		Meta:        arch.Meta,
		TimeCreated: arch.TimeCreated,
		TimeUpdated: arch.TimeUpdated,
		Status:      arch.Status,
		Sources:     arch.Sources,
	}
}

type SourceBucket struct {
	BucketID       string `json:"ID"`
	BucketURL      string
	BucketName     string
	BucketPrefixes []string
	BucketOrigin   string
	Description    string
	Owner          string
}

func (sb SourceBucket) ToBSON() bson.D {
	var prefixes bson.A
	for _, v := range sb.BucketPrefixes {
		prefixes = append(prefixes, v)
	}
	res := bson.D{primitive.E{Key: "_id", Value: sb.BucketID},
		primitive.E{Key: "URL", Value: sb.BucketURL},
		primitive.E{Key: "Name", Value: sb.BucketName},
		primitive.E{Key: "Prefixes", Value: prefixes},
		primitive.E{Key: "Origin", Value: sb.BucketOrigin},
		primitive.E{Key: "Description", Value: sb.Description},
		primitive.E{Key: "Owner", Value: sb.Owner}}

	return res
}

type bucketFileWrapper struct {
	_id     string         `json:"id"`
	buckets []SourceBucket `json:"buckets"`
}

type idFileWrapper struct {
	_id string   `json:"id"`
	Ids []string `json:"ids"`
}
