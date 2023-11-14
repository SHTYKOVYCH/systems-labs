#include "cgi-app.h"

#include <string>
#include <stdexcept>

CGIApp* CGIApp::request(std::string method, std::string path, std::function<void(Http*)> func) {
	paths[method][path] = func;

	return this;
}

void CGIApp::processRequest() {
	try {
		paths[this->http.getMethod()][this->http.getPath()](&this->http);
	}
	catch (std::out_of_range e) {
		this->http.setCode("404", "Not found");
		this->http.write("No such path");
	}

	this->http.addHeader("Content-type: text/plain;");

	this->http.flush();
}
