package logx

type alyslsSettings struct {
	apiDomain    string
	appid        string
	appsecret    string
	projectName  string
	logstoreName string
}

func newAlyslsSettings(settings map[string]interface{}) *alyslsSettings {
	var apiDomain string

	if s1, ok := settings["apiDomain"].(string); ok {
		apiDomain = s1
	}

	var appid string

	if s1, ok := settings["appid"].(string); ok {
		appid = s1
	}

	var appsecret string

	if s1, ok := settings["appsecret"].(string); ok {
		appsecret = s1
	}

	var projectName string

	if s1, ok := settings["projectName"].(string); ok {
		projectName = s1
	}

	var logstoreName string

	if s1, ok := settings["logstoreName"].(string); ok {
		logstoreName = s1
	}

	return &alyslsSettings{
		apiDomain:    apiDomain,
		appid:        appid,
		appsecret:    appsecret,
		projectName:  projectName,
		logstoreName: logstoreName,
	}
}
