
#include <iostream>
#include <string>
#include <vector>
#include <map>
#include <regex>
#include <fstream>
#include <stdio.h>
#include <sstream>
#include <functional>

#include "utils.h"
#include "http.h"
#include "cgi-app.h"

using namespace std;

int main(int argc, char* argv[])
{
	map<string, vector<string>> bd;

	ifstream file("db.bd", fstream::in);

	string new_line;
	while (getline(file, new_line)) {
		vector<string> key_value = split(new_line, ' ');

		for (int i = 1; i < key_value.size(); ++i) {
			bd[key_value[0]].push_back(key_value[i]);
		}
	}
	file.close();

	CGIApp app;

	app.request("GET", "/", [](Http* http) -> void {
		http->setCode("301", "//api");
		});

	app.request("GET", "/api", [&bd](Http* http) -> void {
		http->write("{");

		for (auto needle : bd) {
			http->write("\"" + needle.first + "\": ");

			if (needle.second.size() > 1) {
				http->write("[");
				for (int i = 0; i < needle.second.size(); ++i) {
					http->write("\"" + needle.second[i] + "\"");

					if (i != needle.second.size() - 1) {
						http->write(",");
					}
				}

				http->write("];");
			}
			else {
				http->write("\"" + needle.second[0] + "\";\n");
			}
		}

		http->write("}");
		});

	app.request("POST", "/api", [&bd](Http* http) -> void {
		map<string, vector<string>> body = http->getXwwwHtmlEncodedBody();

		for (auto pair : body) {
			for (auto val : pair.second) {
				bd[pair.first].push_back(val);
			}
		}
		});

	app.request("PATCH", "/api", [&bd](Http* http) -> void {
		map<string, vector<string>> body = http->getXwwwHtmlEncodedBody();

		try {
			for (auto kv : body) {
				bd.at(kv.first);
			}
		}
		catch (out_of_range e) {
			http->setCode("404", "Cannot found one of values");
			return;
		}

		for (auto pair : body) {
			bd.at(pair.first).clear();

			for (auto val : pair.second) {
				bd[pair.first].push_back(val);
			}
		}
		});

	app.request("DELETE", "/api", [&bd](Http* http) -> void {
		map<string, vector<string>> body = http->getXwwwHtmlEncodedBody();

		try {
			for (auto key : body.at("keys")) {
				bd.erase(key);
			}
		}
		catch (out_of_range e) {
			http->setCode("400", "Bad request");
		}
		});

	app.processRequest();

	ofstream ofile("db.bd");

	for (auto it = bd.begin(); it != bd.end(); ++it) {
		ofile << it->first;

		for (auto i : it->second) {
			ofile << ' ' << i;
		}

		ofile << endl;
	}

	ofile.close();

	return -1;
}
