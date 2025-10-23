



const domainGithub = "https://raw.githubusercontent.com/A-440Hz/squids/main/";
// const domainFirebase = "";

interface MediaURLProvider {
    getImageURL(filename: string): string;
}

class GithubURLProvider implements MediaURLProvider {
    getImageURL(filename: string): string {
        return `${domainGithub}/${filename}`;
    }
}

// class FirebaseURLProvider implements MediaURLProvider {
//     getImageURL(filename: string): string {
//         return `${domainFirebase}/${filename}`;
//     }
// }

// TODO: this will be useful if I ever want to have a fallback providers pattern
// https://www.npmjs.com/package/react-image
let defaultMediaURLProvider: MediaURLProvider = new GithubURLProvider();

export function getImageURL(filename: string): string {
    // Optionally add error handling, fallback, etc.
    return defaultMediaURLProvider.getImageURL(filename);
}

// TODO: return this already formatted in json and get rid of this function
export function filenameToTitle(str: string): string {
    return str.split('_').map(w => w[0].toUpperCase() + w.substring(1).toLowerCase()).join(' ');
}
