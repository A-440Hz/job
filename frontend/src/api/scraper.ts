import { scraper_endpoint } from "./endpoint";

const scraperEndpoint = scraper_endpoint + '/scrape';
const summarizeEndpoint = scraper_endpoint + '/summarize';

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

export async function summarizeScrapedData(text: string, itemID: string, model?: string): Promise<any> {
    const req = {text: text, item_id: "", model: ""};
    if (itemID) {
        req.item_id = itemID;
    }
    if (model) {
        req.model = model;
    }
    const res = await fetch(summarizeEndpoint, {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(req)
    });
    if (!res.ok) {
        throw new Error('Failed to summarize scraped data');
    }
    return res.json();
}
