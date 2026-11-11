package com.portfolio.controller;

import com.portfolio.config.AdminCredentialBootstrap;
import com.portfolio.config.AdminSessionFilter;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import jakarta.servlet.http.HttpSession;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.server.ResponseStatusException;

import java.security.SecureRandom;
import java.util.Base64;

@RestController
@RequestMapping("/api/admin")
public class AdminAuthController {
    private static final SecureRandom RANDOM = new SecureRandom();

    public record LoginRequest(String key) {}
    public record SessionResponse(long expiresAt, String csrfToken) {}

    private final AdminCredentialBootstrap credentials;
    private final int sessionMinutes;
    private final boolean secureCookie;

    public AdminAuthController(AdminCredentialBootstrap credentials,
                               @Value("${portfolio.admin.session-minutes:30}") int sessionMinutes,
                               @Value("${server.servlet.session.cookie.secure:false}") boolean secureCookie) {
        this.credentials = credentials;
        this.sessionMinutes = sessionMinutes > 0 ? sessionMinutes : 30;
        this.secureCookie = secureCookie;
    }

    @PostMapping("/login")
    public SessionResponse login(@RequestBody LoginRequest body, HttpServletRequest request,
                                 HttpServletResponse response) {
        response.setHeader("Cache-Control", "no-store");
        if (secureCookie && !request.isSecure()) {
            throw new ResponseStatusException(HttpStatus.FORBIDDEN, "Use HTTPS to log in");
        }
        if (!credentials.configured()) {
            throw new ResponseStatusException(HttpStatus.SERVICE_UNAVAILABLE, "Admin account is not configured");
        }
        if (!credentials.matches(body.key())) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Invalid admin key");
        }

        HttpSession old = request.getSession(false);
        if (old != null) old.invalidate();
        HttpSession session = request.getSession(true);
        long expiresAt = System.currentTimeMillis() + sessionMinutes * 60_000L;
        session.setMaxInactiveInterval(sessionMinutes * 60);
        session.setAttribute(AdminSessionFilter.EXPIRES_AT, expiresAt);
        byte[] random = new byte[32];
        RANDOM.nextBytes(random);
        String csrfToken = Base64.getUrlEncoder().withoutPadding().encodeToString(random);
        session.setAttribute(AdminSessionFilter.CSRF_TOKEN, csrfToken);
        return new SessionResponse(expiresAt, csrfToken);
    }

    @GetMapping("/session")
    public SessionResponse session(HttpSession session) {
        Long expiresAt = (Long) session.getAttribute(AdminSessionFilter.EXPIRES_AT);
        String csrfToken = (String) session.getAttribute(AdminSessionFilter.CSRF_TOKEN);
        if (expiresAt == null || csrfToken == null || System.currentTimeMillis() >= expiresAt) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Admin session expired or missing");
        }
        return new SessionResponse(expiresAt, csrfToken);
    }

    @PostMapping("/logout")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void logout(HttpSession session) {
        session.invalidate();
    }
}
