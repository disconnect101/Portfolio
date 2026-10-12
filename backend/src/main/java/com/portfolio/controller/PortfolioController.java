package com.portfolio.controller;

import com.portfolio.model.Education;
import com.portfolio.model.Experience;
import com.portfolio.model.Profile;
import com.portfolio.model.Project;
import com.portfolio.repository.*;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api")
public class PortfolioController {

    public record PortfolioResponse(Profile profile,
                                    List<Project> projects,
                                    List<Education> education,
                                    List<Experience> experience) {}

    private final ProfileRepository profiles;
    private final ProjectRepository projects;
    private final EducationRepository education;
    private final ExperienceRepository experience;

    public PortfolioController(ProfileRepository profiles, ProjectRepository projects,
                               EducationRepository education, ExperienceRepository experience) {
        this.profiles = profiles;
        this.projects = projects;
        this.education = education;
        this.experience = experience;
    }

    /** Everything the front end needs in one call. */
    @GetMapping("/portfolio")
    public ResponseEntity<PortfolioResponse> portfolio() {
        Profile profile = profiles.findAll().stream().findFirst().orElse(null);
        if (profile == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(new PortfolioResponse(
                profile,
                projects.findAllByOrderByDisplayOrderAsc(),
                education.findAllByOrderByDisplayOrderAsc(),
                experience.findAllByOrderByDisplayOrderAsc()));
    }

    @GetMapping("/profile")
    public ResponseEntity<Profile> profile() {
        return profiles.findAll().stream().findFirst()
                .map(ResponseEntity::ok).orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/projects")
    public List<Project> projects() { return projects.findAllByOrderByDisplayOrderAsc(); }

    @GetMapping("/education")
    public List<Education> education() { return education.findAllByOrderByDisplayOrderAsc(); }

    @GetMapping("/experience")
    public List<Experience> experience() { return experience.findAllByOrderByDisplayOrderAsc(); }
}
