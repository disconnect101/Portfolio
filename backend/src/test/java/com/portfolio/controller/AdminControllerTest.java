package com.portfolio.controller;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.portfolio.config.AdminSessionFilter;
import com.portfolio.model.AdminCredential;
import com.portfolio.repository.AdminCredentialRepository;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.mock.web.MockHttpSession;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.context.TestPropertySource;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.MvcResult;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotEquals;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.delete;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.put;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@SpringBootTest
@AutoConfigureMockMvc
@ActiveProfiles("h2")
@TestPropertySource(properties = "portfolio.admin.bootstrap-key=test-admin-key-with-more-than-32-characters")
class AdminControllerTest {
    private static final String KEY = "test-admin-key-with-more-than-32-characters";

    record Login(MockHttpSession session, String csrfToken) {}

    @Autowired MockMvc mvc;
    @Autowired ObjectMapper json;
    @Autowired AdminCredentialRepository credentials;

    private Login login() throws Exception {
        MvcResult result = mvc.perform(post("/api/admin/login").contentType(MediaType.APPLICATION_JSON)
                        .content("{\"key\":\"" + KEY + "\"}"))
                .andExpect(status().isOk()).andReturn();
        JsonNode response = json.readTree(result.getResponse().getContentAsString());
        return new Login((MockHttpSession) result.getRequest().getSession(false), response.get("csrfToken").asText());
    }

    @Test
    void keyIsHashedInDatabaseAndSessionIsRequired() throws Exception {
        AdminCredential credential = credentials.findById(1L).orElseThrow();
        assertNotEquals(KEY, credential.getPasswordHash());
        assertFalse(credential.getPasswordHash().contains(KEY));

        mvc.perform(get("/api/portfolio")).andExpect(status().isOk());
        mvc.perform(get("/api/admin/session")).andExpect(status().isUnauthorized());
        mvc.perform(post("/api/admin/login").contentType(MediaType.APPLICATION_JSON)
                .content("{\"key\":\"wrong\"}")).andExpect(status().isUnauthorized());
        mvc.perform(post("/api/admin/projects").contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"Should not save\"}")).andExpect(status().isUnauthorized());

        Login login = login();
        mvc.perform(get("/api/admin/session").session(login.session())).andExpect(status().isOk());
        mvc.perform(post("/api/admin/projects").session(login.session())
                .contentType(MediaType.APPLICATION_JSON).content("{\"title\":\"No CSRF\"}"))
                .andExpect(status().isForbidden());

        login.session().setAttribute(AdminSessionFilter.EXPIRES_AT, System.currentTimeMillis() - 1);
        mvc.perform(get("/api/admin/session").session(login.session())).andExpect(status().isUnauthorized());
    }

    @Test
    void canCreateUpdateAndDeleteProject() throws Exception {
        Login login = login();
        String created = mvc.perform(post("/api/admin/projects").session(login.session())
                        .header("X-CSRF-Token", login.csrfToken()).contentType(MediaType.APPLICATION_JSON)
                        .content("{\"title\":\"Test project\",\"technologies\":[\"Java\"],\"displayOrder\":5}"))
                .andExpect(status().isCreated()).andReturn().getResponse().getContentAsString();
        long id = json.readTree(created).get("id").asLong();

        mvc.perform(put("/api/admin/projects/{id}", id).session(login.session())
                        .header("X-CSRF-Token", login.csrfToken()).contentType(MediaType.APPLICATION_JSON)
                        .content("{\"title\":\"Updated project\",\"technologies\":[\"React\"],\"displayOrder\":6}"))
                .andExpect(status().isOk());
        mvc.perform(delete("/api/admin/projects/{id}", id).session(login.session())
                .header("X-CSRF-Token", login.csrfToken())).andExpect(status().isNoContent());
        mvc.perform(delete("/api/admin/projects/{id}", id).session(login.session())
                .header("X-CSRF-Token", login.csrfToken())).andExpect(status().isNotFound());
    }

    @Test
    void canEditProfileAndManageEducationAndExperience() throws Exception {
        Login login = login();
        mvc.perform(put("/api/admin/profile").session(login.session())
                        .header("X-CSRF-Token", login.csrfToken()).contentType(MediaType.APPLICATION_JSON)
                        .content("{\"name\":\"Updated name\",\"title\":\"Engineer\"}"))
                .andExpect(status().isOk());

        for (String section : new String[] {"education", "experience"}) {
            String body = section.equals("education")
                    ? "{\"institution\":\"Test university\",\"displayOrder\":1}"
                    : "{\"company\":\"Test company\",\"highlights\":[\"Shipped a feature\"],\"displayOrder\":1}";
            String created = mvc.perform(post("/api/admin/{section}", section).session(login.session())
                            .header("X-CSRF-Token", login.csrfToken())
                            .contentType(MediaType.APPLICATION_JSON).content(body))
                    .andExpect(status().isCreated()).andReturn().getResponse().getContentAsString();
            long id = json.readTree(created).get("id").asLong();
            mvc.perform(put("/api/admin/{section}/{id}", section, id).session(login.session())
                            .header("X-CSRF-Token", login.csrfToken())
                            .contentType(MediaType.APPLICATION_JSON).content(body))
                    .andExpect(status().isOk());
            mvc.perform(delete("/api/admin/{section}/{id}", section, id).session(login.session())
                    .header("X-CSRF-Token", login.csrfToken())).andExpect(status().isNoContent());
        }
    }

    @Test
    void logoutInvalidatesSession() throws Exception {
        Login login = login();
        mvc.perform(post("/api/admin/logout").session(login.session())
                .header("X-CSRF-Token", login.csrfToken())).andExpect(status().isNoContent());
        mvc.perform(get("/api/admin/session").session(login.session())).andExpect(status().isUnauthorized());
    }
}
