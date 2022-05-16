package main

import "kafka2jenkins/utils"

func init()  {
	utils.JenkinsInit()
	utils.RedisInit()
}

func main() {
	utils.ReadJenkinsTopic()
}
