import { scraper_endpoint } from "./endpoint";

const scraperEndpoint = scraper_endpoint + '/scrape';

export async function fetchScraperData(url: string, itemID: string): Promise<any> {
    const req = {url: url, item_id: ""};
    if (itemID) {
        req.item_id = itemID;
    }
    const res = await fetch(scraperEndpoint, {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(req)
    });
    if (!res.ok) {
        throw new Error('Failed to fetch scraper data');
    }
    return res.json();
}