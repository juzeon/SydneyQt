import Swal, {SweetAlertResult} from 'sweetalert2'

export interface ISwal {
    success: (text: string) => Promise<SweetAlertResult>,
    error: (text: string) => Promise<SweetAlertResult>,
    confirm: (text: string) => Promise<SweetAlertResult>,
}

export let swal: ISwal = {
    success(text: string) {
        return Swal.fire({
            title: 'Success',
            text,
            icon: 'success'
        })
    },
    error(text: string) {
        return Swal.fire({
            title: 'Error',
            text,
            icon: 'error'
        })
    },
    confirm(text:string) {
        return Swal.fire({
            title: 'Confirmation',
            icon: 'question',
            text: text,
            confirmButtonText: 'Confirm',
            cancelButtonText: 'Cancel',
            showCancelButton: true
        })
    }
}
