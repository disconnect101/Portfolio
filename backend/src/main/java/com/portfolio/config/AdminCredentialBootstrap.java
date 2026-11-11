package com.portfolio.config;

import com.portfolio.model.AdminCredential;
import com.portfolio.repository.AdminCredentialRepository;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.stereotype.Component;

import javax.crypto.SecretKeyFactory;
import javax.crypto.spec.PBEKeySpec;
import java.security.MessageDigest;
import java.security.SecureRandom;
import java.util.Base64;

@Component
public class AdminCredentialBootstrap implements ApplicationRunner {
    private static final int ITERATIONS = 600_000;
    private static final SecureRandom RANDOM = new SecureRandom();

    private final AdminCredentialRepository credentials;
    private final String bootstrapKey;

    public AdminCredentialBootstrap(AdminCredentialRepository credentials,
                                    @Value("${portfolio.admin.bootstrap-key:}") String bootstrapKey) {
        this.credentials = credentials;
        this.bootstrapKey = bootstrapKey;
    }

    @Override
    public void run(ApplicationArguments args) {
        if (credentials.existsById(1L) || bootstrapKey.length() < 32) return;
        byte[] salt = new byte[16];
        RANDOM.nextBytes(salt);
        AdminCredential credential = new AdminCredential();
        credential.setId(1L);
        credential.setSalt(Base64.getEncoder().encodeToString(salt));
        credential.setPasswordHash(Base64.getEncoder().encodeToString(derive(bootstrapKey, salt, ITERATIONS)));
        credential.setIterations(ITERATIONS);
        credentials.save(credential);
    }

    public boolean matches(String supplied) {
        if (supplied == null) return false;
        return credentials.findById(1L).map(credential -> {
            byte[] salt = Base64.getDecoder().decode(credential.getSalt());
            byte[] expected = Base64.getDecoder().decode(credential.getPasswordHash());
            byte[] actual = derive(supplied, salt, credential.getIterations());
            return MessageDigest.isEqual(expected, actual);
        }).orElse(false);
    }

    public boolean configured() {
        return credentials.existsById(1L);
    }

    private static byte[] derive(String key, byte[] salt, int iterations) {
        PBEKeySpec spec = new PBEKeySpec(key.toCharArray(), salt, iterations, 256);
        try {
            return SecretKeyFactory.getInstance("PBKDF2WithHmacSHA256").generateSecret(spec).getEncoded();
        } catch (Exception e) {
            throw new IllegalStateException("Could not hash admin key", e);
        } finally {
            spec.clearPassword();
        }
    }
}
