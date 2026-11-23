package com.openjiuwen.demo.professional;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import com.openjiuwen.service.spec.spi.AgentHandler;

@SpringBootApplication
public class ProfessionalAgentApplication {
    public static void main(String[] args) { SpringApplication.run(ProfessionalAgentApplication.class, args); }
    @Bean AgentHandler professionalHandler() { return new ProfessionalHandler(System.getenv()); }
}
