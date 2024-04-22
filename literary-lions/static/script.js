document.addEventListener('DOMContentLoaded', function() {
    // Get the registration modal
    var registrationModal = document.getElementById('registrationModal');
    var registerButton = document.getElementById("registerButton");
    if (registerButton) {
        registerButton.addEventListener('click', openRegistrationModal)
    }
    var closeRegistrationButton = document.querySelector("#registrationModal .close");
    if (closeRegistrationButton) {
        closeRegistrationButton.addEventListener('click', closeRegistrationModal);
    }

    // Function to open the registration modal
    function openRegistrationModal() {
        registrationModal.style.display = "block";
    }

    // Function to close the registration modal
    function closeRegistrationModal() {
        registrationModal.style.display = "none";
    }

    // Close the registration modal when clicking outside of it
    window.onclick = function(event) {
        if (event.target == registrationModal) {
            closeRegistrationModal();
        }
    }

    // Event delegation for opening the posting modal
    document.addEventListener('click', function(event) {
        if (event.target && event.target.id === 'postButton') {
            openPostingModal();
        }
    });

    // Get the posting modal
    var postingModal = document.getElementById('postingModal');
    var closePostingButton = document.querySelector("#postingModal .close");
    if (closePostingButton) {
        closePostingButton.addEventListener('click', closePostingModal);
    }

    // Function to open the posting modal
    function openPostingModal() {
        postingModal.style.display = "block";
    }

    // Function to close the posting modal
    function closePostingModal() {
        postingModal.style.display = "none";
    }

    // Close the posting modal when clicking outside of it
    window.onclick = function(event) {
        if (event.target == postingModal) {
            closePostingModal();
        }
    }
});
