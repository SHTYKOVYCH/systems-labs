
#include <iostream>
#include <string>
#include <vector>
#include <map>
#include <regex>
#include <fstream>
#include <stdio.h>
#include <sstream>
#include <functional>
#include <libpq-fe.h>

#include "utils.h"
#include "http.h"
#include "cgi-app.h"

using namespace std;

int main(int argc, char* argv[])
{
	PGconn* db_connection = PQconnectdb("user=postgres password=postgres host=db-products");
	
	if (PQstatus(db_connection) != CONNECTION_OK)
    {
		cout << "\n\n";
        cout <<  PQerrorMessage(db_connection);
        return -1;
    }	

	CGIApp app;

	PGresult* res;

	app.request("GET", "/", [&res, &db_connection](Http* http) -> void {
		res = PQexec(db_connection, "\
			SELECT * from products; \
		");

		if (PQresultStatus(res) != PGRES_TUPLES_OK) {
			http->setCode("500", "Server error");
			PQclear(res);
			return;
		}
		http->write("[");

		int code_num = PQfnumber(res, "code"), name_num = PQfnumber(res, "name"), weight_num = PQfnumber(res, "weight"), description_num = PQfnumber(res, "description"); 

		for (int i = 0; i < PQntuples(res); ++i) {
			char* code = PQgetvalue(res, i, code_num);
			char* name = PQgetvalue(res, i, name_num);
			char* weight = PQgetvalue(res, i, weight_num);
			char* description = PQgetvalue(res, i, description_num);

			http->write("{");
			http->write("\"code\": \"" + string(code) + "\",");
			http->write("\"name\": \"" + string(name) + "\",");
			http->write("\"weight\": " + string(weight) + ",");
			http->write("\"description\": \"" + string(description) + "\"");
			http->write("}");

			if (i + 1 != PQntuples(res)) {
				http->write(",");
			}
		}

		PQclear(res);

		http->write("]");
		});

	app.request("POST", "/", [&res, &db_connection](Http* http) -> void {
		map<string, string> body = http->getXwwwHtmlEncodedBody();

		try {
			if (body.at("name").size() == 0) {
				throw out_of_range("");
			}

			
			if (body.at("weight").size() == 0) {
				throw out_of_range("");
			}

			
			if (body.at("description").size() == 0) {
				throw out_of_range("");
			}

		} catch (out_of_range e) {
			http->setCode("400", "Bad request");
			return;
		}

		res = PQexec(db_connection, string("\
			INSERT INTO products (name, weight, description) VALUES ('"
				+ body.at("name") + "','"
				+ body.at("weight") + "','"
				+ body.at("description") + "') RETURNING code").c_str());

		if (PQresultStatus(res) != PGRES_TUPLES_OK) {
			http->setCode("500", "Server error");
			PQclear(res);
			return;
		}

		int code_num = PQfnumber(res, "code"); 

		for (int i = 0; i < PQntuples(res); ++i) {
			char* code = PQgetvalue(res, i, code_num);
			http->write(string(code));
		}

		PQclear(res);
		});

	app.request("PATCH", "/", [&res, &db_connection](Http* http) -> void {
		map<string, string> body = http->getXwwwHtmlEncodedBody();

		try {
			if (body.at("code").size() == 0) {
				throw out_of_range("");
			}
			if (body["name"].size() == 0 && body["weight"].size() == 0 && body["description"].size() == 0) {
				return;
			}
		} catch (out_of_range e) {
			http->setCode("400", "Bad request");
			return;
		}

		string cols = "";

		if (body["name"].size() != 0) {
			cols += "name='" + body["name"] + "'";
		}

		if (body["weight"].size() != 0) {
			if (body["name"].size() != 0) {
				cols += ",";
			}			
			cols += "weight='" + body["weight"] + "'";
		}

		if (body["description"].size() != 0) {
			if (body["name"].size() != 0 || body["weight"].size() != 0) {
				cols += ",";
			}
			
			cols += "description='" + body["description"] + "'";
		}


		res = PQexec(db_connection, string("\
			UPDATE products SET " + cols + " WHERE code=" + "'" + body["code"] + "'").c_str());

		if (PQresultStatus(res) != PGRES_COMMAND_OK) {
			http->setCode("500", "Server error");
			PQclear(res);
			return;
		}

		PQclear(res);
		});

	app.request("DELETE", "/", [&res, &db_connection](Http* http) -> void {
		map<string, string> body = http->getXwwwHtmlEncodedBody();

		try {
			if (body.at("code").size() == 0) {
				throw out_of_range("");
			}
		} catch (out_of_range e) {
			http->setCode("400", "Bad request");
			return;
		}

		res = PQexec(db_connection, (string("DELETE FROM products WHERE code=") + string("'") + body["code"] + string("'")).c_str());

		if (PQresultStatus(res) != PGRES_COMMAND_OK) {
			http->setCode("500", "Server error");
			PQclear(res);
			return;
		}

		PQclear(res);
		});

	app.processRequest();

	PQfinish(db_connection);

	return -1;
}
