    import React, { useEffect, useState } from "react";
    import { Pencil, Trash2 } from "lucide-react"; // ไอคอนสวยๆ
    import { Link } from "react-router-dom";

    const ListBookPage = () => {
    const [books, setBooks] = useState([]);

    // ดึงข้อมูลจาก backend
    useEffect(() => {
        fetch("http://localhost:8080/api/v1/books")
        .then((res) => res.json())
        .then((data) => setBooks(data))
        .catch((err) => console.error("Error fetching books:", err));
    }, []);

    // ลบหนังสือ
    const handleDelete = async (id) => {
        if (window.confirm("แน่ใจหรือไม่ว่าต้องการลบหนังสือเล่มนี้?")) {
        try {
            await fetch(`http://localhost:8080/books/${id}`, {
            method: "DELETE",
            });
            setBooks(books.filter((book) => book.id !== id));
        } catch (error) {
            console.error("Error deleting book:", error);
        }
        }
    };
    

    return (
        <div className="p-6">
        <h1 className="text-2xl font-semibold mb-4 text-viridian-700">
            📚 จัดการข้อมูลหนังสือ (Backoffice)
        </h1>

        <table className="min-w-full bg-white border border-gray-200 rounded-lg shadow-md">
            <thead className="bg-viridian-600 text-white">
            <tr>
                <th className="py-2 px-4 text-left">#</th>
                <th className="py-2 px-4 text-left">ชื่อหนังสือ</th>
                <th className="py-2 px-4 text-left">ผู้แต่ง</th>
                <th className="py-2 px-4 text-left">ราคา</th>
                <th className="py-2 px-4 text-left">จัดการ</th>
            </tr>
            </thead>
            <tbody>
            {books.length > 0 ? (
                books.map((book, index) => (
                <tr key={book.id} className="border-b hover:bg-gray-50">
                    <td className="py-2 px-4">{index + 1}</td>
                    <td className="py-2 px-4">{book.title}</td>
                    <td className="py-2 px-4">{book.author}</td>
                    <td className="py-2 px-4">{book.price} บาท</td>
                    <td className="py-2 px-4 flex gap-3">
                    <Link
                        to={`/books/edit/${book.id}`}
                        className="text-blue-600 hover:text-blue-800"
                    >
                        <Pencil className="w-5 h-5" />
                    </Link>
                    <button
                        onClick={() => handleDelete(book.id)}
                        className="text-red-600 hover:text-red-800"
                    >
                        <Trash2 className="w-5 h-5" />
                    </button>
                    </td>
                </tr>
                ))
            ) : (
                <tr>
                <td
                    colSpan="5"
                    className="text-center py-4 text-gray-500 italic"
                >
                    ไม่มีข้อมูลหนังสือ
                </td>
                </tr>
            )}
            </tbody>
        </table>
        </div>
    );
    };

    export default ListBookPage;
