package consts

const (
	MONGO_BASE_URL = "mongodb://39.101.135.5:37017"

	BASE_KAFKA_URL_EXTERNAL = "39.101.180.247:39092"
	BASE_KAFKA_URL_INTERNAL = "10.47.0.1:9092,10.47.0.10:9092,10.47.0.11:9092"

	DEFAULT_KAFKA_PARTITION = 0

	// redis
	REDIS_BASE_SERVER_URL = "39.101.135.5:36379"
	REDIS_KAFKA_TOPIC_OFFSET = "Kafka_Topic_Offset_%s"
	REDIS_HARBOR_REPO_TAG_LIST = "Harbor_TAG_LIST_%s_%s"// Namespace + repo, e.p: Harbor_TAG_LIST_mz2021_passgin, 每一个元素应当为tag的string
	REDIS_HARBOR_REPO_TAG_BUILD = "Harbor_TAG_BUILD_%s_%s_%s" // Namespace + repo + TAG, 存的value为jenkins的build_number e.p: Harbor_TAG_BUILD_mz2021_passgin_1.0

	// jenkins
	JENKINS_KAFKA_TOPIC_BUILD_JOB = "Jenkins_Build_Job"
	JENKINS_DEFAULT_SERVER = "http://39.99.150.232:30002"
	JENKINS_DEFAULT_USER = "root"
	JENKINS_DEFAULT_TOKEN = "11600ce46d61cfcd7e8d97506fbb53f4a1"

	JENKINS_BLUE_OCEAN_BASE_URL = JENKINS_DEFAULT_SERVER + "/blue/organizations/jenkins"
	JENKINS_BLUE_OCEAN_JOB_BUILD_VIEW =  JENKINS_BLUE_OCEAN_BASE_URL + "/%s/detail/%s/%s/pipeline"


)