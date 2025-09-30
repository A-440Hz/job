



const domainGithub = "https://raw.githubusercontent.com/A-440Hz/squids/main/";
const domainFirebase = "";

interface MediaURLProvider {
    getImageURL(filename: string): string;
}

class GithubURLProvider implements MediaURLProvider {
    getImageURL(filename: string): string {
        return `${domainGithub}/${filename}`;
    }
}

class FirebaseURLProvider implements MediaURLProvider {
    getImageURL(filename: string): string {
        return `${domainFirebase}/${filename}`;
    }
}

let defaultMediaURLProvider: MediaURLProvider = new GithubURLProvider();

export function getImageURL(filename: string): string {
    // Optionally add error handling, fallback, etc.
    return defaultMediaURLProvider.getImageURL(filename);
}