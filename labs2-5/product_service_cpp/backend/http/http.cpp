#include "http.h"
#include "utils.h"

#include <iostream>
#include <algorithm>

using namespace std;

Http::Http() {
	if (getenv("QUERY_STRING")) {
		vector<string> paramsAndValues = split(getenv("QUERY_STRING"), '&');

		for (auto pvString : paramsAndValues) {
			vector<string> pvParsed = split(pvString, '=');

			try {
				this->query_params.at(pvParsed[0])->push_back(pvParsed[1]);
			}
			catch (out_of_range e) {
				vector<string>* vec = new vector<string>;

				vec->push_back(pvParsed[1]);

				this->query_params[pvParsed[0]] = vec;
			}
		}
	}

	if (getenv("HTTP_COOKIE")) {
		vector<string> cookieAndValues = split(getenv("HTTP_COOKIE"), ';');

		for (auto cvString : cookieAndValues) {
			vector<string> cv = split(cvString, '=');
			cv[0].erase(cv[0].begin(), std::find_if(cv[0].begin(), cv[0].end(), [](unsigned char ch) {
				return !std::isspace(ch);
				}));
			this->cookie[cv[0]] = cv[1];
		}
	}

	if (getenv("PATH_INFO")) {
		path = getenv("PATH_INFO");
	}
	else {
		path = "/";
	}

	this->body = NULL;

	this->method = "GET";
	this->code = "200";
	this->codeMessage = "Ok";

	if (getenv("REQUEST_METHOD")) {
		this->method = string(getenv("REQUEST_METHOD"));
	}
}

Http::~Http() {
	for (auto param : query_params) {
		delete param.second;
	}

	if (this->body) {
		delete this->body;
	}
}

void Http::flush() {
	cout << "Status: " << code << " " << codeMessage << endl;
	cout << outHeaders;
	cout << "\n\n";
	cout << out;
}

void Http::write(string str) {
	this->out += str;
}

vector<string>* Http::getQueryParam(string name) {
	return this->query_params.at(name);
}

string Http::getRawBody() {
	if (this->body) {
		return *((string*)this->body);
	}
	string retVal = "", line;

	for (std::string line; std::getline(std::cin, line);) {
		retVal += line;
	}

	this->body = new string(retVal);

	return retVal;
}

map<string, string> Http::getXwwwHtmlEncodedBody() {
	string body = this->getRawBody();

	map<string, string> retVal;

	vector<string> namesAndValues = split(body, '&');

	for (auto nvString : namesAndValues) {
		vector<string> nv = split(nvString, '=');

		nv[0].erase(nv[0].begin(), std::find_if(nv[0].begin(), nv[0].end(), [](unsigned char ch) {
			return !std::isspace(ch);
			}));

		retVal[nv[0]] = nv[1];
	}

	return retVal;
}

string Http::getCookie(string name) {
	return this->cookie.at(name);
}

void Http::setCookie(string name, string value) {
	outHeaders += "Set-Cookie: " + name + "=" + value + '\n';
}

void Http::setCode(string code, string message) {
	this->code = code;
	codeMessage = message;
}

void Http::addHeader(string header) {
	outHeaders += header;
}

string Http::getPath() {
	return this->path;
}

string Http::getMethod() {
	return this->method;
}
