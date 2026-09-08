package types

type Template struct {
	TemplateId       int64  `json:"templateId"`
	CampaignId       int64  `json:"campaignId,omitempty"`
	ClientTemplateId string `json:"clientTemplateId,omitempty"`
	CreatedAt        string `json:"createdAt"`
	CreatorUserId    string `json:"creatorUserId"`
	MessageTypeId    int64  `json:"messageTypeId"`
	Name             string `json:"name"`
	UpdatedAt        string `json:"updatedAt"`
}

type TemplatesResponse struct {
	Templates           []Template `json:"templates"`
	NextPageUrl         string     `json:"nextPageUrl,omitempty"`
	PreviousPageUrl     string     `json:"previousPageUrl,omitempty"`
	TotalTemplatesCount int64      `json:"totalTemplatesCount,omitempty"`
}
