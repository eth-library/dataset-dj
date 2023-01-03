package dbutil

import (
	"github.com/eth-library/dataset-dj/util"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// Source contains information about the origin of the data contained within a MetaArchive
type Source struct {
	ID           string `json:"id" bson:"_id"`
	Name         string `json:"name" bson:"name"`
	Organisation string `json:"organisation" bson:"organisation"`
}

type Order struct {
	OrderID   string   `json:"orderID" bson:"_id,omitempty"`
	ArchiveID string   `json:"archiveID" bson:"archiveID"`
	Email     string   `json:"email" bson:"email"`
	Date      string   `json:"date" bson:"date"`
	Status    string   `json:"status" bson:"status"`
	Sources   []string `json:"sources" bson:"sources"`
}

type TimedOrder struct {
	OrderID   string    `json:"orderID" bson:"_id,omitempty"`
	ArchiveID string    `json:"archiveID" bson:"archiveID"`
	Email     string    `json:"email" bson:"email"`
	Date      time.Time `json:"date" bson:"date"`
	Status    string    `json:"status" bson:"status"`
	Sources   []string  `json:"sources" bson:"sources"`
}

type FileGroup struct {
	SourceID string   `json:"sourceID"`
	Files    util.Set `json:"files"`
}

func FileGroupToDB(fg FileGroup) FileGroupDB {
	return FileGroupDB{
		SourceID: fg.SourceID,
		Files:    fg.Files.ToSlice(),
	}
}

type FileGroupDB struct {
	SourceID string   `json:"sourceID" bson:"sourceID"`
	Files    []string `json:"files" bson:"files"`
}

func DBToFileGroup(db FileGroupDB) FileGroup {
	return FileGroup{
		SourceID: db.SourceID,
		Files:    util.SetFromSlice(db.Files),
	}
}

func Union(fgs1 []FileGroup, fgs2 []FileGroup) ([]FileGroup, []string) {
	var contentMap map[string]FileGroup
	for _, fgs := range [][]FileGroup{fgs1, fgs2} {
		contentMap = fillContentMap(fgs, contentMap)
	}
	var res []FileGroup
	var sources []string
	for _, value := range contentMap {
		res = append(res, value)
		sources = append(sources, value.SourceID)
	}
	return res, sources
}

func Unify(fgs []FileGroup) ([]FileGroup, []string) {
	var contentMap = make(map[string]FileGroup)
	contentMap = fillContentMap(fgs, contentMap)
	var res []FileGroup
	var sources []string
	for _, value := range contentMap {
		res = append(res, value)
		sources = append(sources, value.SourceID)
	}
	return res, sources
}

func fillContentMap(fgs []FileGroup, contentMap map[string]FileGroup) map[string]FileGroup {
	for _, fg := range fgs {
		i, ok := contentMap[fg.SourceID]
		if ok {
			contentMap[fg.SourceID] = FileGroup{
				SourceID: i.SourceID,
				Files:    util.SetUnion(fg.Files, i.Files),
			}
		} else {
			contentMap[fg.SourceID] = fg
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
	Sources     []string    `json:"sources"`
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
	ID          string        `json:"id" bson:"_id"`
	Content     []FileGroupDB `json:"content" bson:"content"`
	Meta        string        `json:"meta" bson:"meta"`
	TimeCreated string        `json:"timeCreated" bson:"timeCreated"`
	TimeUpdated string        `json:"timeUpdated" bson:"timeUpdated"`
	Status      string        `json:"status" bson:"status"`
	Sources     []string      `json:"sources" bson:"sources"`
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

type OrderSet struct {
	Elems map[string]Order `json:"elements"`
}

func (os OrderSet) ToSlice() []Order {
	var slice []Order
	for _, o := range os.Elems {
		slice = append(slice, o)
	}
	return slice
}

func (os OrderSet) Add(o Order) {
	os.Elems[o.OrderID] = o
}

type idFileWrapper struct {
	Id  string   `json:"id" bson:"_id"`
	Ids []string `json:"ids" bson:"ids"`
}
