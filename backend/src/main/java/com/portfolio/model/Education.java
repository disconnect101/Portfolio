package com.portfolio.model;

import jakarta.persistence.*;

@Entity
public class Education {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    private String institution;
    private String degree;
    private String fieldOfStudy;
    private String startYear;
    private String endYear;

    @Column(length = 1000)
    private String description;

    private int displayOrder;

    public Education() {}

    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }
    public String getInstitution() { return institution; }
    public void setInstitution(String institution) { this.institution = institution; }
    public String getDegree() { return degree; }
    public void setDegree(String degree) { this.degree = degree; }
    public String getFieldOfStudy() { return fieldOfStudy; }
    public void setFieldOfStudy(String fieldOfStudy) { this.fieldOfStudy = fieldOfStudy; }
    public String getStartYear() { return startYear; }
    public void setStartYear(String startYear) { this.startYear = startYear; }
    public String getEndYear() { return endYear; }
    public void setEndYear(String endYear) { this.endYear = endYear; }
    public String getDescription() { return description; }
    public void setDescription(String description) { this.description = description; }
    public int getDisplayOrder() { return displayOrder; }
    public void setDisplayOrder(int displayOrder) { this.displayOrder = displayOrder; }
}
