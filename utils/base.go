package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/mitchellh/mapstructure"
	"github.com/segmentio/kafka-go"
	"kafka2jenkins/consts"
	"kafka2jenkins/model"
	"strconv"
	"strings"
)

var (
	JenkinsCli = &gojenkins.Jenkins{}
	jenkinsCtx = context.Background()
)

func JenkinsInit() {
	JenkinsCli = gojenkins.CreateJenkins(nil, consts.JENKINS_DEFAULT_SERVER, consts.JENKINS_DEFAULT_USER, consts.JENKINS_DEFAULT_TOKEN)
	_, err := JenkinsCli.Init(jenkinsCtx)
	if err != nil {
		panic(err)
	}

}

func ReadJenkinsTopic()  {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{consts.BASE_KAFKA_URL_EXTERNAL},
		Topic:     consts.JENKINS_KAFKA_TOPIC_BUILD_JOB,
		//Partition: consts.DEFAULT_KAFKA_PARTITION,
		GroupID: "kafka2jenkins",
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB

		// LastOffset 或 FirstOffset, 似乎依然无效
		StartOffset: kafka.LastOffset,
	})
	defer func() {
		_ = r.Close()
	}()

	ctx := context.Background()
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil{
			fmt.Println("[ReadMessage] Err: "+ err.Error())
			break
		}
		// fmt.Printf("\n%s===%s message at offset %d: v=%s", topic, strconv.Itoa(int(goID())), m.Offset, string(m.Value))

		// TODO：如何反馈到前端
		if _, _, err := HandleJenkinsMsg(string(m.Value)); err != nil {
			fmt.Println("[ReadMessage] MongoSet Err: " + err.Error())
			panic(err)
		}
	}
}

func HandleJenkinsMsg(raw string) (int64, string, error) {
	msg := &model.JenkinsKafkaMsg{}
	err := json.Unmarshal([]byte(raw), msg)
	if err != nil {
		return -1, "", err
	}

	params := &model.JenkinsJobBuildParams{}
	err = json.Unmarshal([]byte(msg.JobParams), msg)
	if err != nil {
		return -1, "", err
	}

	job, s, err := BuildJenkinsJob(msg.JobName, params)
	if err != nil {
		return -1, "", err
	}

	return job, s, nil

}

func generateDockerfileForPythonFlask(params *model.JenkinsJobBuildParams) (string, error) {
	sb := strings.Builder{}

	Image := params.Language + ":" + params.LanguageVersion + "-alpine"
	sb.Write([]byte(
		"FROM" + " " + Image + "\n",
	))

	// 复制与工作目录
	sb.Write([]byte(
		"COPY . /usr/src/app/ \n",
	))
	sb.Write([]byte(
		"WORKDIR /usr/src/app/ \n",
	))

	// pip install
	// TODO：gunicorn去掉？
	sb.Write([]byte(
		"RUN pip install --no-cache-dir -r requirements.txt \n",
	))
	sb.Write([]byte(
		"RUN pip install gunicorn==19.6.0 \n",
	))

	// 环境参数
	args := []model.BuildArgKv{}
	err := json.Unmarshal([]byte(params.BuildArgs), &args)
	if err != nil {
		return "", fmt.Errorf("[Params Unmarshal] json.Unmarshal: %s", err.Error())
	}
	for _, arg := range args {
		kv := arg.Key + "=" + arg.Value
		sb.Write([]byte(
			"ENV" + " " + kv + "\n" ,
		))
	}

	// 暴露端口
	ports := strings.Split(params.ExposedPorts, ",")
	for _, port := range ports {
		sb.Write([]byte(
			"EXPOSE" + " " + port + "\n" ,
		))
	}

	// 入口
	sb.Write([]byte(
		"ENTRYPOINT [\"/usr/local/bin/gunicorn\"] \n",
	))

	// 参数
	ipPort := "0.0.0.0:" + ports[0]
	sb.Write([]byte(
		fmt.Sprintf("CMD [\"-w\",\"1\",\"-b\",\"%s\",\"app:app\"]", ipPort),
	))

	fmt.Println(sb.String())

	return sb.String(), nil
}

func generateDockerfileForGolangGin(params *model.JenkinsJobBuildParams) (string, error) {
	sb := strings.Builder{}

	// 采用镜像， TODO： 修改常量
	Image := "alpine"
	sb.Write([]byte(
		"FROM" + " " + Image + "\n",
	))

	// 复制
	sb.Write([]byte(
		"ADD . /go/src/app \n",
	))

	// 工作目录
	sb.Write([]byte(
		"WORKDIR /go/src/app \n",
	))

	// test
	sb.Write([]byte(
		"RUN pwd \n",
	))
	sb.Write([]byte(
		"RUN ls -l \n",
	))

	//path := params.CompiledProductPath
	//sb.Write([]byte(
	//	"WORKDIR" + " " + path + "\n",
	//))

	// 环境参数
	args := []model.BuildArgKv{}
	err := json.Unmarshal([]byte(params.BuildArgs), &args)
	if err != nil {
		return "", fmt.Errorf("[Params Unmarshal] json.Unmarshal: %s", err.Error())
	}
	for _, arg := range args {
		kv := arg.Key + "=" + arg.Value
		sb.Write([]byte(
			"ENV" + " " + kv + "\n" ,
		))
	}

	// 暴露端口
	ports := strings.Split(params.ExposedPorts, ",")
	for _, port := range ports {
		sb.Write([]byte(
			"EXPOSE" + " " + port + "\n" ,
		))
	}

	// 容器入口
	entryPoint := "[\"" + params.EntryPoint + "\"]"
	sb.Write([]byte(
		"ENTRYPOINT" + " " + entryPoint + "\n" ,
	))

	fmt.Println(sb.String())
	return sb.String(), nil
}

func BuildJenkinsJob(jobName string, params *model.JenkinsJobBuildParams) (int64, string, error) {
	buildParams := map[string]string{}
	err := mapstructure.Decode(params, &buildParams)
	if err != nil {
		fmt.Printf("\n[BuildJenkinsJob Failed] mapstructure Decode: %s\n", err.Error())
		return -1, "", fmt.Errorf("\n[BuildJenkinsJob Failed] mapstructure Decode: %s\n", err.Error())
	}

	// 回退，下面这段删掉，去掉后面注释
	strDockerfile := ""
	switch {
	case strings.Contains(jobName, "ginex"):
		strDockerfile, err = generateDockerfileForGolangGin(params)
	case strings.Contains(jobName, "kitex"):
	case strings.Contains(jobName, "flask"):
		strDockerfile, err = generateDockerfileForPythonFlask(params)
	}
	if err != nil {
		fmt.Printf("\n[BuildJenkinsJob Failed] generateDockerfileForGolangGin Err: %s\n", err.Error())
		return -1, "", fmt.Errorf("\n[BuildJenkinsJob Failed] generateDockerfileForGolangGin Err: %s\n", err.Error())
	}
	buildParams["StrDockerfile"] = strDockerfile
	buildParams["CompiledProductPath"] = params.CompiledProductPath

	job, err := JenkinsCli.GetJob(jenkinsCtx, jobName, []string{}...)
	if err != nil {
		fmt.Printf("\n[BuildJenkinsJob Failed] JenkinsCli GetJob: %s\n", err.Error())
		return -1, "", fmt.Errorf("\n[BuildJenkinsJob Failed] JenkinsCli GetJob: %s\n", err.Error())
	}

	queueid, err := job.InvokeSimple(jenkinsCtx, buildParams)
	if err != nil {
		fmt.Printf("\n[BuildJenkinsJob Failed] job InvokeSimple: %s\n", err.Error())
		return -1, "", fmt.Errorf("\n[BuildJenkinsJob Failed] job InvokeSimple: %s\n", err.Error())
	}

	build, err := JenkinsCli.GetBuildFromQueueID(jenkinsCtx, queueid)
	if err != nil {
		fmt.Printf("\n[BuildJenkinsJob Failed] JenkinsCli GetBuildFromQueueID: %s\n", err.Error())
		return -1, "", fmt.Errorf("\n[BuildJenkinsJob Failed] JenkinsCli GetBuildFromQueueID: %s\n", err.Error())
	}


	blueOceanUrl := fmt.Sprintf(
		consts.JENKINS_BLUE_OCEAN_JOB_BUILD_VIEW,
		jobName, jobName, strconv.FormatInt(build.GetBuildNumber(), 10),
	)

	// Redis保存构建历史
	jobBuildKey := fmt.Sprintf(consts.REDIS_HARBOR_REPO_TAG_BUILD, params.HarborNs, params.Psm, params.ImageTag)
	err = RedisSetString(jobBuildKey, jobName + "#" + strconv.FormatInt(build.GetBuildNumber(), 10))
	if err != nil {
		fmt.Printf("\n[BuildJenkinsJob Failed] RedisSetInt: %s\n", err.Error())
	}
	tagListKey := fmt.Sprintf(consts.REDIS_HARBOR_REPO_TAG_LIST, params.HarborNs, params.Psm)
	err = RedisSetStringSet(tagListKey, []string{params.ImageTag})
	if err != nil {
		fmt.Printf("\n[BuildJenkinsJob Failed] RedisSetStringSet: %s\n", err.Error())
	}

	return build.GetBuildNumber(), blueOceanUrl, nil
}
