package main

type OrderRequestBody struct {
	Sources []string `json:"sources"`
}

type OrderStatusBody struct {
	NewStatus string `json:"newStatus"`
}

type Archive struct {
	ID          string      `json:"id"`
	Content     []FileGroup `json:"content"`
	Meta        string      `json:"meta"`
	TimeCreated string      `json:"timeCreated"`
	TimeUpdated string      `json:"timeUpdated"`
	Status      string      `json:"status"`
	Sources     []string    `json:"sources"`
}

type FileGroup struct {
	SourceID string   `json:"sourceID"`
	Files    []string `json:"files"`
}

type Secrets struct {
	HandlerKEY       string `yaml:"handlerKEY"`
	MailUser         string `yaml:"mailUser"`
	MailPassword     string `yaml:"mailPassword"`
	LibDriveUser     string `yaml:"libDriveUser"`
	LibDrivePassword string `yaml:"libDrivePassword"`
}

type ApplicationConfig struct {
	Sources     []Source `yaml:"sources"`
	ApiURL      string   `yaml:"apiURL"`
	MailHost    string   `yaml:"mailHost"`
	MailAddress string   `yaml:"mailAddress"`
}

type Source struct {
	SourceID        string `yaml:"sourceID"`
	SourceName      string `yaml:"sourceName"`
	SourceStartTime string `yaml:"sourceStartTime"`
	SourceEndTime   string `yaml:"sourceEndTime"`
	SourceFileHost  string `yaml:"sourceFileHost"`
	SourceHostType  string `yaml:"sourceHostType"`
}

type LibDriveConfig struct {
	Host       string `yaml:"host"`
	ApiPath    string `yaml:"apiPath"`
	WebdavPath string `yaml:"webdavPath"`
}
