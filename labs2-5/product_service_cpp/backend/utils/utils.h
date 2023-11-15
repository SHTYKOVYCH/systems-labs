#pragma once

#include <vector>
#include <string>
#include <sstream>

std::vector<std::string> split(const std::string& s, char delimiter);

void ltrim(std::string& s);
void rtrim(std::string& s);
void trim(std::string& s);