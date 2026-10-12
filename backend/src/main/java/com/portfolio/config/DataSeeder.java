package com.portfolio.config;

import com.portfolio.model.Education;
import com.portfolio.model.Experience;
import com.portfolio.model.Profile;
import com.portfolio.model.Project;
import com.portfolio.repository.*;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * Seeds PLACEHOLDER content on first start (only when the profile table is empty).
 * Edit the values below with your real details, then wipe the tables (or the docker volume)
 * and restart, or just update the rows directly in the database.
 */
@Component
public class DataSeeder implements CommandLineRunner {

    private final ProfileRepository profiles;
    private final ProjectRepository projects;
    private final EducationRepository education;
    private final ExperienceRepository experience;

    public DataSeeder(ProfileRepository profiles, ProjectRepository projects,
                      EducationRepository education, ExperienceRepository experience) {
        this.profiles = profiles;
        this.projects = projects;
        this.education = education;
        this.experience = experience;
    }

    @Override
    public void run(String... args) {
        if (profiles.count() > 0) return;

        Profile p = new Profile();
        p.setName("Aditya");
        p.setTitle("Software Engineer");
        p.setDescription("I'm a developer who enjoys building clean, reliable software and turning ideas into "
                + "polished products. This is placeholder text: replace it with a short paragraph about who you are, "
                + "what you work on, and what you care about.");
        p.setCoverPhotoUrl("/cover.svg");
        p.setProfilePhotoUrl("/avatar.svg");
        p.setEmail("aditya.dev.forall@gmail.com");
        p.setLocation("India");
        p.setGithubUrl("https://github.com/your-username");
        p.setLinkedinUrl("https://linkedin.com/in/your-username");
        profiles.save(p);

        projects.saveAll(List.of(
                project(1, "Project One", "Short description of what this project does and the problem it solves.",
                        List.of("Java", "Spring Boot", "PostgreSQL"), "https://github.com/your-username/project-one", null),
                project(2, "Project Two", "Short description of another project. Mention your role and the outcome.",
                        List.of("React", "TypeScript", "Vite"), "https://github.com/your-username/project-two", "https://example.com"),
                project(3, "Project Three", "Short description of a third project, such as a 3D model or a side project.",
                        List.of("Blender", "3D Printing"), null, null)));

        education.saveAll(List.of(
                edu(1, "Your University", "Bachelor of Technology", "Computer Science", "2018", "2022",
                        "Relevant coursework, honors, or activities."),
                edu(2, "Your School", "Higher Secondary", "Science", "2016", "2018", null)));

        experience.saveAll(List.of(
                exp(1, "Company Name", "Software Engineer", "Bengaluru, India", "Jul 2022", "Present",
                        List.of("Built and shipped features used by thousands of users.",
                                "Improved API response times by refactoring slow queries.",
                                "Mentored junior developers and reviewed code.")),
                exp(2, "Previous Company", "Software Engineering Intern", "Remote", "Jan 2022", "Jun 2022",
                        List.of("Developed internal tooling in Java.",
                                "Wrote unit and integration tests."))));
    }

    private Project project(int order, String title, String desc, List<String> tech, String repo, String live) {
        Project x = new Project();
        x.setDisplayOrder(order);
        x.setTitle(title);
        x.setDescription(desc);
        x.setTechnologies(new java.util.ArrayList<>(tech));
        x.setRepoUrl(repo);
        x.setLiveUrl(live);
        return x;
    }

    private Education edu(int order, String inst, String degree, String field, String start, String end, String desc) {
        Education e = new Education();
        e.setDisplayOrder(order);
        e.setInstitution(inst);
        e.setDegree(degree);
        e.setFieldOfStudy(field);
        e.setStartYear(start);
        e.setEndYear(end);
        e.setDescription(desc);
        return e;
    }

    private Experience exp(int order, String company, String role, String loc, String start, String end, List<String> highlights) {
        Experience e = new Experience();
        e.setDisplayOrder(order);
        e.setCompany(company);
        e.setRole(role);
        e.setLocation(loc);
        e.setStartDate(start);
        e.setEndDate(end);
        e.setHighlights(new java.util.ArrayList<>(highlights));
        return e;
    }
}
