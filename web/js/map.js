// Initialize the Leaflet map
(function() {
    'use strict';

    // Default location (London, UK) - will be updated by Go
    const DEFAULT_LAT = 51.5074;
    const DEFAULT_LON = -0.1278;
    const DEFAULT_ZOOM = 13;

    // Create the map
    window.map = L.map('map', {
        center: [DEFAULT_LAT, DEFAULT_LON],
        zoom: DEFAULT_ZOOM,
        zoomControl: true,
        attributionControl: true
    });

    // Add OpenStreetMap tile layer
    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 19,
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(window.map);

    // Handle map click events
    window.map.on('click', function(e) {
        const lat = e.latlng.lat;
        const lon = e.latlng.lng;

        // Update marker position visually
        setMarker(lat, lon);

        // Notify Go about the click
        notifyMapClick(lat, lon);
    });

    // Handle map load complete
    window.map.on('load', function() {
        hideLoading();
    });

    // Handle tile loading errors
    window.map.on('tileerror', function(e) {
        console.warn('Tile loading error:', e);
    });

    // Hide loading overlay
    function hideLoading() {
        const loading = document.getElementById('loading');
        if (loading) {
            loading.classList.add('hidden');
            setTimeout(function() {
                loading.style.display = 'none';
            }, 300);
        }
    }

    // Initialize WebChannel bridge and notify Go when ready
    initBridge()
        .then(function(bridge) {
            console.log('Bridge initialized, map ready');
            hideLoading();
            notifyMapReady();
        })
        .catch(function(error) {
            console.error('Failed to initialize bridge:', error);
            hideLoading();
            // Map still works, just without Go communication
        });

    // Hide loading after a timeout in case events don't fire
    setTimeout(hideLoading, 3000);

    // Expose map instance globally for debugging
    window.goldenHourMap = {
        map: window.map,
        setMarker: setMarker,
        centerMap: centerMap,
        setLocation: setLocation,
        clearMarker: clearMarker,
        getMapBounds: getMapBounds
    };

})();
