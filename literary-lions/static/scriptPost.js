document.addEventListener('DOMContentLoaded', function() {
        // Event delegation for opening the comment modal
    document.addEventListener('click', function(event) {
        if (event.target && event.target.id === 'commentButton') {
            openCommentModal();
        }
    });

    // Get the comment modal
    var commentModal = document.getElementById('commentModal');
    var closeCommentButton = document.querySelector("#commentModal .close");
    if (closeCommentButton) {
        closeCommentButton.addEventListener('click', closeCommentModal);
    }

    // Function to open the comment modal
    function openCommentModal() {
        commentModal.style.display = "block";
    }

    // Function to close the comment modal
    function closeCommentModal() {
        commentModal.style.display = "none";
    }

    // Close the comment modal when clicking outside of it
    window.onclick = function(event) {
        if (event.target == commentModal) {
            closeCommentModal();
        }
    }
});
