#pragma once

#include <functional>

#include "http.h"

class CGIApp {
private:
	Http http;
	std::map<std::string, std::map<std::string, std::function<void(Http*)>>> paths;
public:
	CGIApp* request(std::string method, std::string path, std::function<void(Http*)> func);

	void processRequest();
};