package com.portfolio.model;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;

@Entity
public class AdminCredential {
    @Id
    private Long id;

    @Column(nullable = false)
    private String salt;

    @Column(nullable = false)
    private String passwordHash;

    private int iterations;

    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }
    public String getSalt() { return salt; }
    public void setSalt(String salt) { this.salt = salt; }
    public String getPasswordHash() { return passwordHash; }
    public void setPasswordHash(String passwordHash) { this.passwordHash = passwordHash; }
    public int getIterations() { return iterations; }
    public void setIterations(int iterations) { this.iterations = iterations; }
}
