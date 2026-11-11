package com.portfolio.controller;

import com.portfolio.model.Education;
import com.portfolio.model.Experience;
import com.portfolio.model.Profile;
import com.portfolio.model.Project;
import com.portfolio.repository.EducationRepository;
import com.portfolio.repository.ExperienceRepository;
import com.portfolio.repository.ProfileRepository;
import com.portfolio.repository.ProjectRepository;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

import java.util.ArrayList;

@RestController
@RequestMapping("/api/admin")
public class AdminController {

    private final ProfileRepository profiles;
    private final ProjectRepository projects;
    private final EducationRepository education;
    private final ExperienceRepository experience;

    public AdminController(ProfileRepository profiles, ProjectRepository projects,
                           EducationRepository education, ExperienceRepository experience) {
        this.profiles = profiles;
        this.projects = projects;
        this.education = education;
        this.experience = experience;
    }

    @PutMapping("/profile")
    public Profile updateProfile(@RequestBody Profile input) {
        Profile profile = profiles.findAll().stream().findFirst().orElseGet(Profile::new);
        profile.setName(required(input.getName(), "Name"));
        profile.setTitle(input.getTitle());
        profile.setDescription(input.getDescription());
        profile.setCoverPhotoUrl(input.getCoverPhotoUrl());
        profile.setProfilePhotoUrl(input.getProfilePhotoUrl());
        profile.setEmail(input.getEmail());
        profile.setLocation(input.getLocation());
        profile.setGithubUrl(input.getGithubUrl());
        profile.setLinkedinUrl(input.getLinkedinUrl());
        return profiles.save(profile);
    }

    @PostMapping("/projects")
    @ResponseStatus(HttpStatus.CREATED)
    public Project createProject(@RequestBody Project input) {
        return projects.save(copyProject(new Project(), input));
    }

    @PutMapping("/projects/{id}")
    public Project updateProject(@PathVariable Long id, @RequestBody Project input) {
        return projects.save(copyProject(projects.findById(id).orElseThrow(AdminController::notFound), input));
    }

    @DeleteMapping("/projects/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteProject(@PathVariable Long id) {
        if (!projects.existsById(id)) throw notFound();
        projects.deleteById(id);
    }

    @PostMapping("/education")
    @ResponseStatus(HttpStatus.CREATED)
    public Education createEducation(@RequestBody Education input) {
        return education.save(copyEducation(new Education(), input));
    }

    @PutMapping("/education/{id}")
    public Education updateEducation(@PathVariable Long id, @RequestBody Education input) {
        return education.save(copyEducation(education.findById(id).orElseThrow(AdminController::notFound), input));
    }

    @DeleteMapping("/education/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteEducation(@PathVariable Long id) {
        if (!education.existsById(id)) throw notFound();
        education.deleteById(id);
    }

    @PostMapping("/experience")
    @ResponseStatus(HttpStatus.CREATED)
    public Experience createExperience(@RequestBody Experience input) {
        return experience.save(copyExperience(new Experience(), input));
    }

    @PutMapping("/experience/{id}")
    public Experience updateExperience(@PathVariable Long id, @RequestBody Experience input) {
        return experience.save(copyExperience(experience.findById(id).orElseThrow(AdminController::notFound), input));
    }

    @DeleteMapping("/experience/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteExperience(@PathVariable Long id) {
        if (!experience.existsById(id)) throw notFound();
        experience.deleteById(id);
    }

    private static Project copyProject(Project target, Project input) {
        target.setTitle(required(input.getTitle(), "Title"));
        target.setDescription(input.getDescription());
        target.setTechnologies(input.getTechnologies() == null ? new ArrayList<>() : new ArrayList<>(input.getTechnologies()));
        target.setImageUrl(input.getImageUrl());
        target.setLiveUrl(input.getLiveUrl());
        target.setRepoUrl(input.getRepoUrl());
        target.setDisplayOrder(input.getDisplayOrder());
        return target;
    }

    private static Education copyEducation(Education target, Education input) {
        target.setInstitution(required(input.getInstitution(), "Institution"));
        target.setDegree(input.getDegree());
        target.setFieldOfStudy(input.getFieldOfStudy());
        target.setStartYear(input.getStartYear());
        target.setEndYear(input.getEndYear());
        target.setDescription(input.getDescription());
        target.setDisplayOrder(input.getDisplayOrder());
        return target;
    }

    private static Experience copyExperience(Experience target, Experience input) {
        target.setCompany(required(input.getCompany(), "Company"));
        target.setRole(input.getRole());
        target.setLocation(input.getLocation());
        target.setStartDate(input.getStartDate());
        target.setEndDate(input.getEndDate());
        target.setHighlights(input.getHighlights() == null ? new ArrayList<>() : new ArrayList<>(input.getHighlights()));
        target.setDisplayOrder(input.getDisplayOrder());
        return target;
    }

    private static String required(String value, String label) {
        if (value == null || value.isBlank()) {
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST, label + " is required");
        }
        return value.trim();
    }

    private static ResponseStatusException notFound() {
        return new ResponseStatusException(HttpStatus.NOT_FOUND, "Item not found");
    }
}
