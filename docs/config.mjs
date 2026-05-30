const stage = process.env.NODE_ENV || "dev"
const isProduction = stage === "production"

export default {
  url: isProduction ? "https://vandordev.github.io" : "http://localhost:4321",
  basePath:  isProduction ? "/vx" : "/",
  github: "https://github.com/vandordev/vx",
  githubDocs: "https://github.com/vandordev/vx",
  title: "vx",
  description: "A modern terminal-first CLI from Vandor Dev, built with Go, Cobra, and Bubble Tea.",
}
