// Caution! Be sure you understand the caveats before publishing an application with
// offline support. See https://aka.ms/blazor-offline-considerations

self.importScripts('./service-worker-assets.js');
self.importScripts('./js/idb.js');

self.addEventListener('install', event => event.waitUntil(onInstall(event)));
self.addEventListener('activate', event => event.waitUntil(onActivate(event)));
self.addEventListener('fetch', event => event.respondWith(onFetch(event)));

const cacheNamePrefix = 'resource-cache-';
const apiCachePrefix = "api-cache-";
const imageCacheName = "image-cache";
const cardImageCacheName = "card-image-cache";
const cacheName = `${cacheNamePrefix}${self.assetsManifest.version}`;
const apiCacheName = `${apiCachePrefix}${self.assetsManifest.version}`;
const offlineAssetsInclude = [ /\.dll$/, /\.pdb$/, /\.wasm/, /\.html/, /\.js$/, /\.css$/, /\.png$/, /\.jpeg$/, /\.jpg$/, /\.gif$/, /\.webp$/, /\.svg$/, /\.mp3$/, /\.wav$/, /\.json$/, /\.webmanifest$/ ];
const offlineAssetsExclude = [ /service-worker\.js$/, /app\.json$/, ];

async function onInstall(event) {
    self.skipWaiting();
    const assetsRequests = self.assetsManifest.assets
        .filter(asset => offlineAssetsInclude.some(pattern => pattern.test(asset.url)))
        .filter(asset => !offlineAssetsExclude.some(pattern => pattern.test(asset.url)))
        .map(asset => {
            return new Request(asset.url, {
                cache: "reload",
            });
        });
	for (const request of assetsRequests){
		await caches.open(cacheName).then(cache => cache.add(request)).catch(error => {
			console.error("Failed to cache:", request, error);
            somethingFailed = true;
		});
	}
    await prepOutbox();
}

async function onActivate(event) {
    const cacheKeys = await caches.keys();
    await Promise.all(cacheKeys
        .filter(key => key.startsWith(cacheNamePrefix) && key !== cacheName)
        .map(key => caches.delete(key)));
    await Promise.all(cacheKeys
        .filter(key => key.startsWith(apiCachePrefix) && key !== cacheName)
        .map(key => caches.delete(key)));
    await Promise.all(cacheKeys
        .filter(key => key.startsWith(imageCacheName) && key !== cacheName)
        .map(key => caches.delete(key)));
    await Promise.all(cacheKeys
        .filter(key => key.startsWith(cardImageCacheName) && key !== cacheName)
        .map(key => caches.delete(key)));
    reloadClients(true);
}

async function tryAppCache(request){
    const cache = await caches.open(cacheName);
    const cachedResponse = await cache.match(request);
    return cachedResponse;
}

async function tryImageCache(request){
    const cache = await caches.open(imageCacheName);
    const cachedResponse = await cache.match(request);
    return cachedResponse;
}

async function tryCardImageCache(request){
    const cache = await caches.open(cardImageCacheName);
    const cachedResponse = await cache.match(request);
    return cachedResponse;
}

async function tryFetch(request){
    const response =  await fetch(request);
    // Skip caching bad responses
    if (!response || response.status !== 200 || response.type !== "basic" && response.type !== "cors" || response.redirected) {
        return response;
    }
    // Only cache image API responses
    if (response.type === "cors"){
        const responseToCache = response.clone();
        if (response.url.indexOf("/image") !== -1){
            if (request.url.indexOf("/card/" !== -1)){
                const imgCache = await caches.open(cardImageCacheName);
                await imgCache.put(request, responseToCache);
            } else {
                const imgCache = await caches.open(imageCacheName);
                await imgCache.put(request, responseToCache);
            }
        } else if (response.url.indexOf("https://api." !== -1) || response.url.indexOf("http://api." !== -1)) {
            const apiCache = await caches.open(apiCacheName);
            await apiCache.put(request, responseToCache);
        }
    } else if (response.type === "basic"){
        await appCache.put(request, responseToCache);
    }
    return response;
}

async function onFetch(event) {
    const shouldServeIndexHtml = event.request.mode === 'navigate';
    const request = shouldServeIndexHtml ? 'index.html' : event.request;
    try {
        if (event.request.method === 'GET' && !event.request.url.match(/app\.json$/)) {
            let response = await tryAppCache(request);
            if (!response){
                if (request.url.indexOf("/image") !== -1){
                    if (request.url.indexOf("/card/" !== -1)){
                        response = await tryCardImageCache(request);
                        if (response){
                            return response;
                        }
                    } else {
                        response = await tryImageCache(request);
                        if (response){
                            return response;
                        }
                    }
                }
            }
            if (!response){
                response = await tryFetch(request);
            }
            return response;
        } else {
            return fetch(request);
        }
    } catch (e){
        // API cache is only hit when the client doesn't have a network connection
        const apiCache = await caches.open(apiCacheName);
        const cachedResponse = await apiCache.match(request);
        if (cachedResponse){
            return cachedResponse;
        } else {
            throw "Network error";
        }
    }
}

function reloadClients(force = false){
	self.clients.matchAll().then(clients => {
		clients.forEach(client => {
			if (!client.focused || force){
				client.postMessage({
					type: "reload",
				});
			}
		});
	});
}

let db = null;

function clearCache(){
    return new Promise(async (resolve) => {
        caches.delete(imageCacheName);
        caches.delete(apiCacheName);
        indexedDB.deleteDatabase("service-worker");
        const request = await fetch(`${self.origin}/schema.json`);
        const response = await request.json();
        const db = await idb.openDB("application", response.version, {
            blocked() {
                indexedDB.deleteDatabase("applicaiton");
                resolve();
            },
            blocking() {
                indexedDB.deleteDatabase("applicaiton");
                resolve();
            },
        });
        const tables = response.tables;
        for (let t = 0; t < tables.length; t++){
            const name = tables[t].name;
            const key = tables[t]?.keyPath ?? "id";
            if (typeof tables[t]?.preserve === "undefined" || tables[t]?.preserve !== true){
                const rows = await db.getAll(name);
                for (const row of rows){
                    await db.delete(name, row[key]);
                }
            }
        }
        const ingestRows = await db.getAll("ingest-tracker");
        for (const row of ingestRows){
            await db.delete("ingest-tracker", row["route"]);
        }
        resolve();
    });
}

function queue(url, method, payload){
    const data = {
        uid: Date.now(),
        url: url,
        method: method,
        payload: null,
    };
    if (typeof payload === "object" && payload !== null){
        data.payload = JSON.stringify(payload);
    } else if (payload !== null){
        data.payload = payload;
    }
    db.put("outbox", data);
}

async function prepOutbox(){
    db = await idb.openDB("service-worker", 1, {
        upgrade(db, oldVersion, newVersion, transaction) {
            // Purge old stores so we don't brick the service worker when upgrading
            for (let i = 0; i < db.objectStoreNames.length; i++) {
                db.deleteObjectStore(db.objectStoreNames[i]);
            }
            const outbox = db.createObjectStore("outbox", {
                keyPath: "uid",
                autoIncrement: false,
            });
            outbox.createIndex("uid", "uid", { unique: true });
            outbox.createIndex("url", "url", { unique: false });
            outbox.createIndex("method", "method", { unique: false });
            outbox.createIndex("payload", "payload", { unique: false });
        },
    });
}

async function tryRequest(request){
    try {
        const response = await fetch(request.url, {
            method: request.method,
            credentials: "include",
            headers: new Headers({
                "Content-Type": "application/json",
            }),
            body: request.payload,
        });
        return response.ok;
    } catch (e){
        return false;
    }
}

let flushingOutbox = false;
async function flushOutbox(){
    if(flushingOutbox){
        return;
    }
    flushingOutbox = true;
    if (db === null){
        await prepOutbox();
    }
    const requests = await db.getAll("outbox");
    for (const request of requests){
        if (navigator.onLine){
            const success = await tryRequest(request);
            if (success){
                await db.delete("outbox", request.uid);
            } else {
                break;
            }
        } else {
            break;
        }
    }
    flushingOutbox = false;
}

self.onmessage = async (event) => {
    const { type } = event.data;
    switch (type){
        case "flush-outbox":
            flushOutbox();
            break;
        case "queue":
            if (event.data?.url && event.data?.method){
                queue(event.data.url, event.data.method, event.data.payload);
            }
            break;
		case "login":
			reloadClients();
			break;
		case "logout":
            await clearCache();
			reloadClients();
			break;
        default:
            break;
    }
}

