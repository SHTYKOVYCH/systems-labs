#pragma once

#include <string>
#include <map>
#include <vector>

class Http {
private:
	std::map<std::string, std::vector<std::string>*> query_params;
	std::map<std::string, std::string> cookie;
	std::string* body;

	std::string code;
	std::string codeMessage;
	std::string outHeaders;
	std::string out;
	std::string path;
	std::string method;

public:
	Http();

	~Http();

	void flush();

	void write(std::string str);

	std::vector<std::string>* getQueryParam(std::string name);

	std::string getRawBody();

	std::map<std::string, std::string> getXwwwHtmlEncodedBody();

	std::string getCookie(std::string name);

	void setCookie(std::string name, std::string value);

	void setCode(std::string code, std::string message);

	void addHeader(std::string header);

	std::string getPath();

	std::string getMethod();
};
