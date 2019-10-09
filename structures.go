package main

type Image struct {
	StartDate string `json:"startdate"`
	FullStartDate string `json:"fullstartdate"`
	EndDate string `json:"enddate"`
	Url string `json:"url"`
	UrlBase string `json:"urlbase"`
	Copyright string `json:"copyright"`
	Title string `json:"title"`
	Quiz string `json:"quiz"`
	Hash string `json:"hsh"`
}

type Images struct {
	Images []struct {Image}
}