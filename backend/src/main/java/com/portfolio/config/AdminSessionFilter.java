package com.portfolio.config;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import jakarta.servlet.http.HttpSession;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;

@Component
public class AdminSessionFilter extends OncePerRequestFilter {
    public static final String EXPIRES_AT = "adminExpiresAt";
    public static final String CSRF_TOKEN = "adminCsrfToken";

    @Override
    protected boolean shouldNotFilter(HttpServletRequest request) {
        String path = request.getRequestURI().substring(request.getContextPath().length());
        return !path.startsWith("/api/admin/") || path.equals("/api/admin/login");
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response,
                                    FilterChain filterChain) throws ServletException, IOException {
        response.setHeader("Cache-Control", "no-store");
        HttpSession session = request.getSession(false);
        Long expiresAt = session == null ? null : (Long) session.getAttribute(EXPIRES_AT);
        if (expiresAt == null || System.currentTimeMillis() >= expiresAt) {
            if (session != null) session.invalidate();
            response.sendError(HttpServletResponse.SC_UNAUTHORIZED, "Admin session expired or missing");
            return;
        }

        String method = request.getMethod();
        if (!method.equals("GET") && !method.equals("HEAD") && !method.equals("OPTIONS")) {
            String expected = (String) session.getAttribute(CSRF_TOKEN);
            String supplied = request.getHeader("X-CSRF-Token");
            if (expected == null || supplied == null || !MessageDigest.isEqual(
                    expected.getBytes(StandardCharsets.UTF_8), supplied.getBytes(StandardCharsets.UTF_8))) {
                response.sendError(HttpServletResponse.SC_FORBIDDEN, "Invalid CSRF token");
                return;
            }
        }

        filterChain.doFilter(request, response);
    }
}
