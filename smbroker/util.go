package smbroker

//GetServiceNamesByType to get the service names for the given type
func GetServiceNamesByType(services []Service, serviceType ServiceType) []string {
	var srvNames []string
	for _, v := range services {
		//if serviceType is empty , then all the services
		if v.Type == serviceType || serviceType == ServiceTypeAll {
			srvNames = append(srvNames, v.Name)
		}
	}
	return nil
}
