import 'htmx.org';
import 'htmx-ext-json-enc';
import Swal from 'sweetalert2';

document.addEventListener('htmx:confirm', function(evt) {
  // Fancy delete confirmation popup
  if (evt.target.matches("[confirm-delete='true']")) {
    evt.preventDefault();
    Swal.fire({
      title: "Are you sure?",
      text: "Are you sure you want to delete this task?",
      icon: "warning",
      showCancelButton: true,
    }).then((result) => {
      if (result.isConfirmed) {
        evt.detail.issueRequest();
      }
    });
  }
});

document.addEventListener('click', function(evt) {
  // Clear filters functionality
  if (evt.target.closest('.btn-clear-filters')) {
    const nameFilter = document.getElementById('name-filter');
    const completionFilter = document.getElementById('completion-filter');

    if (nameFilter) nameFilter.value = '';
    if (completionFilter) completionFilter.value = '';
  }
});
