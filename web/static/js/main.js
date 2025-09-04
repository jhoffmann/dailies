import 'htmx.org';
import 'htmx-ext-json-enc';
import Swal from 'sweetalert2/dist/sweetalert2.all.min.js';

document.addEventListener('htmx:confirm', function(evt) {
  if (evt.target.matches("[confirm-with-sweet-alert='true']")) {
    evt.preventDefault();
    Swal.fire({
      title: "Are you sure?",
      text: "Are you sure you are sure?",
      icon: "warning",
      showCancelButton: true,
      dangerMode: true,
    }).then((result) => {
      if (result.isConfirmed) {
        evt.detail.issueRequest();
      }
    });
  }
});
