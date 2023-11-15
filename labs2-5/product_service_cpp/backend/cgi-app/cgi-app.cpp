#include "cgi-app.h"

#include <string>
#include <stdexcept>

CGIApp* CGIApp::request(std::string method, std::string path, std::function<void(Http*)> func) {
	paths[path][method] = func;

	return this;
}

void CGIApp::processRequest() {
	try {
		paths.at(this->http.getPath());
	} catch (std::out_of_range e) {
		this->http.setCode("404", "Not found");
	}
	try {
		paths.at(this->http.getPath()).at(this->http.getMethod())(&this->http);
	}
	catch (std::out_of_range e) {
		this->http.setCode("405", "Method not allowed");
	}

	this->http.addHeader("Content-type: text/plain;");

	this->http.flush();
}
