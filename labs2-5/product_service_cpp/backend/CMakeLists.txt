cmake_minimum_required (VERSION 3.8)

project ("lab1")

find_package(PostgreSQL REQUIRED)

include_directories("./http" "./utils" "./cgi-app")
include_directories(${PostgreSQL_INCLUDE_DIRS})

set(HEADERS
        utils/utils.h
        http/http.h
        cgi-app/cgi-app.h
        )

set(SOURCES
        utils/utils.cpp
        http/http.cpp
        cgi-app/cgi-app.cpp
        lab1.cpp      
        )

# Добавьте источник в исполняемый файл этого проекта.
add_executable (lab1 ${HEADERS} ${SOURCES})
target_link_libraries(lab1 -static-libgcc -static-libstdc++)
target_link_libraries(lab1 ${PostgreSQL_LIBRARIES})

