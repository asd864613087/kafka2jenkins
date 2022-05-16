package model

type JenkinsJobBuildResp struct {
	BuildNumber  int    `json:"build_number"`
	BlueOceanUrl string `json:"blue_ocean_url"`
}

type JenkinsJobBuildParams struct {
	GitRepo      string `json:"git_repo"`
	GitBranch    string `json:"git_branch"`
	ImageTag     string `json:"image_tag"`
	JobToken     string `json:"job_token"`
	BuildScript  string `json:"build_script"`
	BuildArgs    string `json:"build_args"`
	ExposedPorts string `json:"exposed_ports"`
	EntryPoint   string `json:"entry_point"`

	CompiledProductPath string `json:"compiled_product_path"`
	Psm                 string `json:"psm"`
	HarborNs            string `json:"harbor_ns"`

	Language        string `json:"language"`
	LanguageVersion string `json:"language_version"`
	LanguageFramework string `json:"language_framework"`
}

type BuildArgKv struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type JenkinsJobs struct {
	Jobs []*JenkinsJob
}

type JenkinsJob struct {
	Name     string            `json:"name"`
	Builds   []JenkinsJobBuild `json:"jenkins_job_build"`
	URL      string            `json:"url"`
	BuildCnt int               `json:"build_cnt"`
}

type JenkinsJobBuild struct {
	Number int64  `json:"number"`
	URL    string `json:"url"`
}

type JenkinsKafkaMsg struct {
	JobName string `json:"job_name"`
	JobParams string `json:"job_params"`
}
