import React, { useEffect, useState } from "react";
import { Pencil, Trash2, PlusCircle } from "lucide-react";
import { Link } from "react-router-dom";

const ListBookPage = () => {
  const [books, setBooks] = useState([]);

  useEffect(() => {
    fetch("http://localhost:8080/api/v1/books")
      .then((res) => res.json())
      .then((data) => setBooks(data))
      .catch((err) => console.error("Error fetching books:", err));
  }, []);

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
    <div className="p-8 bg-gradient-to-br from-emerald-50 via-white to-emerald-100 min-h-screen">
      <div className="flex justify-between items-center mb-6 border-b border-emerald-300 pb-3">
        <h1 className="text-3xl font-bold text-emerald-700 drop-shadow-sm">
          📚 จัดการข้อมูลหนังสือ (Backoffice)
        </h1>
        <Link
          to="/store-manager/add-book"
          className="flex items-center gap-2 bg-emerald-600 hover:bg-emerald-700 text-white px-5 py-2.5 rounded-xl shadow-md transition duration-200 hover:scale-105"
        >
          <PlusCircle className="w-5 h-5" />
          <span className="font-medium">เพิ่มหนังสือ</span>
        </Link>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full bg-white border border-gray-200 rounded-xl shadow-lg overflow-hidden">
          <thead className="bg-emerald-600 text-white text-sm uppercase">
            <tr>
              <th className="py-3 px-4 text-left">#</th>
              <th className="py-3 px-4 text-left">ชื่อหนังสือ</th>
              <th className="py-3 px-4 text-left">ผู้แต่ง</th>
              <th className="py-3 px-4 text-left">ราคา</th>
              <th className="py-3 px-4 text-left">จัดการ</th>
            </tr>
          </thead>
          <tbody className="text-gray-700">
            {books.length > 0 ? (
              books.map((book, index) => (
                <tr
                  key={book.id}
                  className="border-b hover:bg-emerald-50 transition duration-150"
                >
                  <td className="py-3 px-4">{index + 1}</td>
                  <td className="py-3 px-4 font-medium">{book.title}</td>
                  <td className="py-3 px-4">{book.author}</td>
                  <td className="py-3 px-4">{book.price} บาท</td>
                  <td className="py-3 px-4 flex gap-3">
                    <Link
                      to={`/books/edit/${book.id}`}
                      className="text-blue-600 hover:text-blue-800 transition transform hover:scale-110"
                    >
                      <Pencil className="w-5 h-5" />
                    </Link>
                    <button
                      onClick={() => handleDelete(book.id)}
                      className="text-red-600 hover:text-red-800 transition transform hover:scale-110"
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
                  className="text-center py-6 text-gray-500 italic bg-gray-50"
                >
                  ไม่มีข้อมูลหนังสือ
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default ListBookPage;
