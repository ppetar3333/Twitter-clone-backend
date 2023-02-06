FROM nginx:1.18

COPY ./api_gateway.conf /etc/nginx/conf.d/api_gateway.conf
COPY ./example.com.crt /etc/ssl/certs/example.com.crt
COPY ./example.com.csr /etc/ssl/csr/example.com.csr
COPY ./example.com.key /etc/ssl/private/example.com.key
COPY ./example.com.key.org /etc/ssl/keyorg/example.com.key.org
COPY ./v3.ext /etc/ssl/v3/v3.ext
RUN rm /etc/nginx/conf.d/default.conf

#  RUN apt-get update ; apt-get install unzip
#  ADD https://github.com/open-telemetry/opentelemetry-cpp-contrib/releases/download/webserver%2Fv1.0.0/opentelemetry-webserver-sdk-x64-linux.tgz.zip /opt
#  RUN cd /opt ; unzip opentelemetry-webserver-sdk-x64-linux.tgz.zip; tar xvfz opentelemetry-webserver-sdk-x64-linux.tgz
#  RUN cd /opt/opentelemetry-webserver-sdk; ./install.sh
#  ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/opt/opentelemetry-webserver-sdk/sdk_lib/lib
#  RUN echo "load_module /opt/opentelemetry-webserver-sdk/WebServerModule/Nginx/ngx_http_opentelemetry_module.so;\n$(cat /etc/nginx/nginx.conf)" > /etc/nginx/nginx.conf
#  COPY opentelemetry_module.conf /etc/nginx/conf.d

