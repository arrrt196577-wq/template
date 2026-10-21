package com.example.template;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.server.LocalServerPort;
import org.springframework.context.ApplicationContext;
import org.springframework.security.core.userdetails.UserDetailsService;

import static org.assertj.core.api.Assertions.assertThat;

@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
class ApplicationSmokeTest {

    @LocalServerPort
    private int port;

    @Autowired
    private ApplicationContext context;

    private final HttpClient client = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .followRedirects(HttpClient.Redirect.NEVER)
            .build();

    @Test
    void healthIsPublicWithoutSessionOrInfrastructureDetails() throws Exception {
        var response = get("/actuator/health");

        assertThat(response.statusCode()).isEqualTo(200);
        assertThat(response.body()).contains("\"status\":\"UP\"").doesNotContain("components");
        assertThat(response.headers().allValues("set-cookie")).isEmpty();
    }

    @Test
    void nonHealthPathsAreDeniedWithoutLoginRedirectOrSession() throws Exception {
        for (var path : new String[] {"/", "/api/v1/users", "/login", "/actuator/env"}) {
            var response = get(path);

            assertThat(response.statusCode()).as(path).isEqualTo(403);
            assertThat(response.headers().allValues("set-cookie")).isEmpty();
            assertThat(response.headers().allValues("location")).isEmpty();
            assertThat(response.headers().allValues("www-authenticate")).isEmpty();
        }
    }

    @Test
    void bearerTextDoesNotEnableAccessBeforeTokenAuthenticationExists() throws Exception {
        var request = request("/api/v1/users")
                .header("Authorization", "Bearer placeholder-not-a-real-token")
                .GET()
                .build();

        var response = client.send(request, HttpResponse.BodyHandlers.ofString());

        assertThat(response.statusCode()).isEqualTo(403);
        assertThat(response.headers().allValues("set-cookie")).isEmpty();
    }

    @Test
    void writesAreDeniedWithoutCreatingASession() throws Exception {
        var response = client.send(request("/api/v1/users")
                        .POST(HttpRequest.BodyPublishers.ofString("{}"))
                        .header("Content-Type", "application/json")
                        .build(),
                HttpResponse.BodyHandlers.ofString());

        assertThat(response.statusCode()).isEqualTo(403);
        assertThat(response.headers().allValues("set-cookie")).isEmpty();
    }

    @Test
    void bootstrapDoesNotCreateDefaultInMemoryUsers() {
        assertThat(context.getBeansOfType(UserDetailsService.class)).isEmpty();
    }

    private HttpResponse<String> get(String path) throws Exception {
        return client.send(request(path).GET().build(), HttpResponse.BodyHandlers.ofString());
    }

    private HttpRequest.Builder request(String path) {
        return HttpRequest.newBuilder(URI.create("http://127.0.0.1:" + port + path))
                .timeout(Duration.ofSeconds(5));
    }
}
