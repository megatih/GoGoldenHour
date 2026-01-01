// WebChannel bridge to communicate between JavaScript and Go
let goBridge = null;
let bridgeReady = false;

// Initialize the WebChannel connection to Go
function initBridge() {
    return new Promise((resolve, reject) => {
        // Check if we're running inside Qt WebEngine
        if (typeof qt === 'undefined' || typeof qt.webChannelTransport === 'undefined') {
            console.warn('Qt WebChannel not available - running in standalone mode');
            // Create a mock bridge for development/testing
            goBridge = {
                onMapClick: function(lat, lon) {
                    console.log('Mock bridge: onMapClick', lat, lon);
                },
                onMapReady: function() {
                    console.log('Mock bridge: onMapReady');
                },
                onError: function(error) {
                    console.error('Mock bridge: onError', error);
                }
            };
            bridgeReady = true;
            resolve(goBridge);
            return;
        }

        new QWebChannel(qt.webChannelTransport, function(channel) {
            goBridge = channel.objects.goBridge;
            if (goBridge) {
                bridgeReady = true;
                console.log('WebChannel connected to Go bridge');
                resolve(goBridge);
            } else {
                reject(new Error('Go bridge object not found in WebChannel'));
            }
        });
    });
}

// Check if the bridge is ready
function isBridgeReady() {
    return bridgeReady && goBridge !== null;
}

// Notify Go when user clicks on the map
function notifyMapClick(lat, lon) {
    if (isBridgeReady() && goBridge.onMapClick) {
        goBridge.onMapClick(lat, lon);
    }
}

// Notify Go when the map is fully loaded
function notifyMapReady() {
    if (isBridgeReady() && goBridge.onMapReady) {
        goBridge.onMapReady();
    }
}

// Notify Go of any errors
function notifyError(errorMessage) {
    if (isBridgeReady() && goBridge.onError) {
        goBridge.onError(errorMessage);
    }
    console.error('Map error:', errorMessage);
}

// ============================================
// Functions called FROM Go to update the map
// ============================================

// Set or update the marker position
function setMarker(lat, lon) {
    if (window.currentMarker) {
        window.currentMarker.setLatLng([lat, lon]);
    } else {
        // Create a custom golden-hour themed marker
        const goldenIcon = L.divIcon({
            className: 'golden-hour-marker',
            iconSize: [20, 20],
            iconAnchor: [10, 10]
        });
        window.currentMarker = L.marker([lat, lon], { icon: goldenIcon }).addTo(window.map);
    }
    return true;
}

// Center the map on specific coordinates
function centerMap(lat, lon, zoom) {
    if (window.map) {
        const z = zoom || window.map.getZoom() || 13;
        window.map.setView([lat, lon], z);
        return true;
    }
    return false;
}

// Set marker and center map in one operation
function setLocation(lat, lon, zoom) {
    setMarker(lat, lon);
    centerMap(lat, lon, zoom);
    return true;
}

// Show a popup at the marker location
function showPopup(lat, lon, content) {
    if (window.currentMarker) {
        window.currentMarker.bindPopup(content).openPopup();
    } else {
        L.popup()
            .setLatLng([lat, lon])
            .setContent(content)
            .openOn(window.map);
    }
    return true;
}

// Clear the current marker
function clearMarker() {
    if (window.currentMarker) {
        window.map.removeLayer(window.currentMarker);
        window.currentMarker = null;
    }
    return true;
}

// Get current map bounds (for potential future use)
function getMapBounds() {
    if (window.map) {
        const bounds = window.map.getBounds();
        return {
            north: bounds.getNorth(),
            south: bounds.getSouth(),
            east: bounds.getEast(),
            west: bounds.getWest()
        };
    }
    return null;
}
