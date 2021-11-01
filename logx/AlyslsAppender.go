package logx

import (
	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/gogo/protobuf/proto"
	"github.com/meiguonet/mgboot-go-common/enum/DatetimeFormat"
	"regexp"
	"strings"
	"time"
)

type alyslsAppender struct {
}

func (a *alyslsAppender) GetAppenderName() string {
	return "SlyslsAppender"
}

func (a *alyslsAppender) Write(buf []byte) (int, error) {
	st := globalAlyslsSettings

	if st.appid == "" || st.appsecret == "" || st.projectName == "" || st.logstoreName == "" {
		return len(buf), nil
	}

	parts := strings.Split(string(buf), fieldSep)
	var ts, channel, level, msg string

	for _, p := range parts {
		if strings.HasPrefix(p, "ts:") {
			ts = strings.TrimPrefix(p, "ts:")
			continue
		}

		if strings.HasPrefix(p, "channel:") {
			channel = strings.TrimPrefix(p, "channel:")
			continue
		}

		if strings.HasPrefix(p, "level:") {
			level = strings.TrimPrefix(p, "level:")
			continue
		}

		if strings.HasPrefix(p, "msg:") {
			msg = strings.TrimPrefix(p, "msg:")
		}
	}

	msgTags, msg := a.handleMsgTags(msg)

	contents := []*sls.LogContent{
		{Key: proto.String("ts"), Value: proto.String(ts)},
		{Key: proto.String("channel"), Value: proto.String(channel)},
		{Key: proto.String("level"), Value: proto.String(level)},
		{Key: proto.String("message"), Value: proto.String(msg)},
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	t1, _ := time.ParseInLocation(DatetimeFormat.Full, ts, loc)

	logItem := &sls.Log{
		Time:     proto.Uint32(uint32(t1.Unix())),
		Contents: contents,
	}

	logs := []*sls.Log{logItem}
	projectName := st.projectName
	storeName := st.logstoreName

	tags := []*sls.LogTag{
		{Key: proto.String("project"), Value: proto.String(projectName)},
		{Key: proto.String("store"), Value: proto.String(storeName)},
	}

	for _, parts := range msgTags {
		tags = append(tags, &sls.LogTag{
			Key:   proto.String(parts[0]),
			Value: proto.String(parts[1]),
		})
	}

	group := &sls.LogGroup{Logs: logs, LogTags: tags}

	client := sls.CreateNormalInterface(
		st.apiDomain,
		st.appid,
		st.appsecret,
		"",
	)

	client.PutLogs(projectName, storeName, group)
	client.Close()
	return len(buf), nil
}

func (a *alyslsAppender) handleMsgTags(msg string) ([][]string, string) {
	re1 := regexp.MustCompile(`^tags[\x20\t]*=[\x20\t]*\[([^]]+)]`)
	tags := make([][]string, 0)
	groups := re1.FindStringSubmatch(msg)

	if len(groups) < 2 {
		return tags, strings.TrimSpace(msg)
	}

	msg = strings.TrimPrefix(msg, groups[0])
	msg = strings.TrimSpace(msg)
	re2 := regexp.MustCompile(`[\x20\t]+`)
	parts := re2.Split(strings.TrimSpace(groups[1]), -1)
	re3 := regexp.MustCompile(`[\x20\t]*:[\x20\t]*`)

	for _, p := range parts {
		a1 := re3.Split(p, -1)

		if len(a1) != 2 {
			continue
		}

		p1 := strings.TrimSpace(a1[0])
		p2 := strings.TrimSpace(a1[1])

		if p1 == "" || p2 == "" {
			continue
		}

		tags = append(tags, []string{p1, p2})
	}

	return tags, msg
}
