import 'htmx.org';
import 'htmx-ext-json-enc';
import _hyperscript from 'hyperscript.org';
import Swal from 'sweetalert2';

_hyperscript.browserInit();

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


