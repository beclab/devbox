package server

type InstallationResponseData struct {
	UID string `json:"uid"`
}
type Response struct {
	Code int32 `json:"code"`
}

type InstallationResponse struct {
	Response
	Data InstallationResponseData `json:"data"`
}

type SystemServerWrap struct {
	Code    int32                `json:"code"`
	Message string               `json:"message"`
	Data    InstallationResponse `json:"data"`
}

type App struct {
	Title string `json:"title"`
}

type RenameApp struct {
	Name string `json:"name"`
}
