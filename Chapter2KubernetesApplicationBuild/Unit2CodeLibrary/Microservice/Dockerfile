# Use the official Traefik image as the base image
FROM traefik:v2.9.10

# Copy traefik.yml and dynamic.yml to the container
COPY traefik.yml /etc/traefik/traefik.yml
COPY dynamic.yml /etc/traefik/dynamic.yml

# Expose the port that Traefik will be listening on
EXPOSE 80

# Start Traefik with the specified configuration file
CMD ["traefik", "--configFile=/etc/traefik/traefik.yml"]
