package com.portfolio.repository;

import com.portfolio.model.AdminCredential;
import org.springframework.data.jpa.repository.JpaRepository;

public interface AdminCredentialRepository extends JpaRepository<AdminCredential, Long> {
}
