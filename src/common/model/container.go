package model

type Container struct {
	Name       string `json:"name"`
	WorkingDir string `json:"working_dir"`
	//VolumeMounts  VolumeStruct `json:"volume_mount"`
	VolumeMounts  []VolumeMountStruct `json:"volume_mounts"`
	Image         ImageIndex          `json:"image"`
	Env           []EnvStruct         `json:"env"`
	ContainerPort []int               `json:"container_port"`
	Command       string              `json:"command"`
	CPURequest    string              `json:"cpu_request"`
	MemRequest    string              `json:"mem_request"`
	CPULimit      string              `json:"cpu_limit"`
	MemLimit      string              `json:"mem_limit"`
}

type VolumeStruct struct {
	TargetStorageService string `json:"target_storage_service"`
	TargetPath           string `json:"target_path"`
	VolumeName           string `json:"volume_name"`
	ContainerPath        string `json:"container_path"`
	//mount type: 0, folder; 1, file
	MountTypeFlag int `json:"mount_type_flag"`
}

type VolumeMountStruct struct {
	VolumeType    string `json:"volume_type"`
	VolumeName    string `json:"volume_name"`
	ContainerPath string `json:"container_path"`
	//mount type: 0, folder; 1, file
	ContainerPathFlag    int    `json:"container_path_flag"`
	TargetStorageService string `json:"target_storage_service"`
	TargetPath           string `json:"target_path"`
	TargetPVC            string `json:"target_pvc"`
}
